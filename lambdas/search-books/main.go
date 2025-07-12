package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// GoogleBooksResponse represents the response from Google Books API
type GoogleBooksResponse struct {
	TotalItems int           `json:"totalItems"`
	Items      []BookItem    `json:"items"`
}

// BookItem represents a single book item from Google Books API
type BookItem struct {
	ID         string     `json:"id"`
	VolumeInfo VolumeInfo `json:"volumeInfo"`
}

// VolumeInfo contains the book information
type VolumeInfo struct {
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	ImageLinks *ImageLinks `json:"imageLinks,omitempty"`
}

// ImageLinks contains book cover image URLs
type ImageLinks struct {
	Thumbnail string `json:"thumbnail"`
}

// SearchResult represents the simplified response we return
type SearchResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the search query from query parameters
	query, queryOK := request.QueryStringParameters["q"]
	if !queryOK || query == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Missing required query parameter 'q'"}`,
		}, nil
	}

	// Build the Google Books API URL
	googleBooksURL := fmt.Sprintf("https://www.googleapis.com/books/v1/volumes?q=%s&maxResults=10", url.QueryEscape(query))

	// Make the HTTP request to Google Books API
	resp, err := http.Get(googleBooksURL)
	if err != nil {
		log.Printf("Error calling Google Books API: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to search for books"}`,
		}, nil
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading Google Books API response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to process search response"}`,
		}, nil
	}

	// Parse the Google Books response
	var googleResponse GoogleBooksResponse
	if err := json.Unmarshal(body, &googleResponse); err != nil {
		log.Printf("Error parsing Google Books API response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to parse search response"}`,
		}, nil
	}

	// Transform the response to our simplified format
	searchResults := make([]SearchResult, 0, len(googleResponse.Items))
	for _, item := range googleResponse.Items {
		result := SearchResult{
			ID:    item.ID,
			Title: item.VolumeInfo.Title,
		}

		// Join authors into a single string
		if len(item.VolumeInfo.Authors) > 0 {
			result.Author = item.VolumeInfo.Authors[0]
			if len(item.VolumeInfo.Authors) > 1 {
				for _, author := range item.VolumeInfo.Authors[1:] {
					result.Author += ", " + author
				}
			}
		}

		// Add thumbnail if available
		if item.VolumeInfo.ImageLinks != nil {
			result.Thumbnail = item.VolumeInfo.ImageLinks.Thumbnail
		}

		searchResults = append(searchResults, result)
	}

	// Marshal the results
	responseBody, err := json.Marshal(searchResults)
	if err != nil {
		log.Printf("Error marshalling search results: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to format search results"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	// If the LAMBDA_TASK_ROOT environment variable is not set, we're running locally.
	if os.Getenv("LAMBDA_TASK_ROOT") == "" {
		// Create a dummy request for local testing.
		request := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"q": "the hobbit",
			},
		}

		// Call the handler directly.
		response, err := handler(request)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}

		// Print the response body to stdout.
		fmt.Println("--- Local execution ---")
		// pretty print json
		var prettyJSON []map[string]interface{}
		json.Unmarshal([]byte(response.Body), &prettyJSON)
		prettyBody, _ := json.MarshalIndent(prettyJSON, "", "  ")
		fmt.Println(string(prettyBody))

	} else {
		// Start the Lambda handler in the AWS environment.
		lambda.Start(handler)
	}
}