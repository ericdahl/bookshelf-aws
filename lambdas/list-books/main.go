package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Book represents a book record.
type Book struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Author string   `json:"author"`
	Status string   `json:"status"`
	Rating int      `json:"rating,omitempty"`
	Review string   `json:"review,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// For this initial version, we return a static list of books.
	// In a real application, you would fetch this from a database like DynamoDB,
	// likely using the user's identity from the request context to fetch their books.
	books := []Book{
		{
			ID:     "book_01",
			Title:  "The Hobbit",
			Author: "J.R.R. Tolkien",
			Status: "reading",
			Rating: 5,
			Tags:   []string{"fantasy", "classic"},
		},
		{
			ID:     "book_02",
			Title:  "Dune",
			Author: "Frank Herbert",
			Status: "finished",
			Rating: 5,
			Tags:   []string{"sci-fi", "classic"},
		},
		{
			ID:     "book_03",
			Title:  "Project Hail Mary",
			Author: "Andy Weir",
			Status: "want-to-read",
		},
	}

	body, err := json.Marshal(books)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}, nil
}

func main() {
	// If the LAMBDA_TASK_ROOT environment variable is not set, we're running locally.
	if os.Getenv("LAMBDA_TASK_ROOT") == "" {
		// Create a dummy request for local testing.
		request := events.APIGatewayProxyRequest{}

		// Call the handler directly.
		response, err := handler(request)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}

		// Print the response body to stdout.
		fmt.Println("--- Local execution ---")
		fmt.Println(response.Body)
	} else {
		// Start the Lambda handler in the AWS environment.
		lambda.Start(handler)
	}
} 