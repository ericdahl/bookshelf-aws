package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
)

const tableName = "books"

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

// BookRequest represents the request payload for creating a book.
type BookRequest struct {
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Series     string   `json:"series,omitempty"`
	Status     string   `json:"status"`
	Rating     *int     `json:"rating,omitempty"`
	Review     string   `json:"review,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	StartedAt  string   `json:"started_at,omitempty"`
	FinishedAt string   `json:"finished_at,omitempty"`
	Thumbnail  string   `json:"thumbnail,omitempty"`
}

// Book represents a book record for DynamoDB.
type Book struct {
	PK         string   `dynamodbav:"PK"`
	SK         string   `dynamodbav:"SK"`
	ID         string   `dynamodbav:"id"`
	Title      string   `dynamodbav:"Title"`
	Author     string   `dynamodbav:"Author"`
	Series     string   `dynamodbav:"Series,omitempty"`
	Status     string   `dynamodbav:"status"`
	Rating     *int     `dynamodbav:"rating,omitempty"`
	Review     string   `dynamodbav:"review,omitempty"`
	Tags       []string `dynamodbav:"tags,omitempty"`
	StartedAt  string   `dynamodbav:"started_at,omitempty"`
	FinishedAt string   `dynamodbav:"finished_at,omitempty"`
	Thumbnail  string   `dynamodbav:"thumbnail,omitempty"`
}

// APIBook is the structure for the API response.
type APIBook struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Series     string   `json:"series,omitempty"`
	Status     string   `json:"status"`
	Rating     *int     `json:"rating,omitempty"`
	Review     string   `json:"review,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	StartedAt  string   `json:"started_at,omitempty"`
	FinishedAt string   `json:"finished_at,omitempty"`
	Thumbnail  string   `json:"thumbnail"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse the request body
	var bookRequest BookRequest
	if err := json.Unmarshal([]byte(request.Body), &bookRequest); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	// Validate required fields
	if bookRequest.Title == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Title is required",
		}, nil
	}

	if bookRequest.Author == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Author is required",
		}, nil
	}

	// Validate status
	validStatuses := map[string]bool{
		"WANT_TO_READ": true,
		"READING":      true,
		"READ":         true,
	}
	if bookRequest.Status == "" {
		bookRequest.Status = "WANT_TO_READ" // Default status
	} else if !validStatuses[bookRequest.Status] {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid status. Must be one of: WANT_TO_READ, READING, READ",
		}, nil
	}

	// Generate UUID for the book
	bookID := uuid.New().String()

	// Create the book record
	book := Book{
		PK:         "BOOK#" + bookID,
		SK:         "BOOK",
		ID:         bookID,
		Title:      bookRequest.Title,
		Author:     bookRequest.Author,
		Series:     bookRequest.Series,
		Status:     bookRequest.Status,
		Rating:     bookRequest.Rating,
		Review:     bookRequest.Review,
		Tags:       bookRequest.Tags,
		StartedAt:  bookRequest.StartedAt,
		FinishedAt: bookRequest.FinishedAt,
		Thumbnail:  bookRequest.Thumbnail,
	}

	// Marshal the book to DynamoDB attributes
	item, err := attributevalue.MarshalMap(book)
	if err != nil {
		log.Printf("Error marshalling book to DynamoDB attributes: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Put the item in DynamoDB
	tableNameVar := tableName
	putInput := &dynamodb.PutItemInput{
		TableName: &tableNameVar,
		Item:      item,
	}

	_, err = ddbClient.PutItem(context.TODO(), putInput)
	if err != nil {
		log.Printf("Error putting item to DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Create the API response
	apiBook := APIBook{
		ID:         book.ID,
		Title:      book.Title,
		Author:     book.Author,
		Series:     book.Series,
		Status:     book.Status,
		Rating:     book.Rating,
		Review:     book.Review,
		Tags:       book.Tags,
		StartedAt:  book.StartedAt,
		FinishedAt: book.FinishedAt,
		Thumbnail:  book.Thumbnail,
	}

	body, err := json.Marshal(apiBook)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
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
		request := events.APIGatewayProxyRequest{
			Body: `{
				"title": "Test Book",
				"author": "Test Author",
				"series": "Test Series",
				"status": "WANT_TO_READ",
				"rating": 5,
				"review": "Great book!",
				"tags": ["fantasy", "adventure"]
			}`,
		}

		// Call the handler directly.
		response, err := handler(request)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}

		// Print the response details to stdout.
		fmt.Println("--- Local execution ---")
		fmt.Printf("Status Code: %d\n", response.StatusCode)
		if response.StatusCode == http.StatusCreated {
			// Pretty print JSON
			var prettyJSON map[string]interface{}
			json.Unmarshal([]byte(response.Body), &prettyJSON)
			prettyBody, _ := json.MarshalIndent(prettyJSON, "", "  ")
			fmt.Println(string(prettyBody))
		} else {
			fmt.Printf("Response Body: %s\n", response.Body)
		}
	} else {
		// Start the Lambda handler in the AWS environment.
		lambda.Start(handler)
	}
} 