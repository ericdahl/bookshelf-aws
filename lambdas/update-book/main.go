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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "books"

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

// BookUpdateRequest represents the request payload for updating a book.
type BookUpdateRequest struct {
	Title      *string  `json:"title,omitempty"`
	Author     *string  `json:"author,omitempty"`
	Series     *string  `json:"series,omitempty"`
	Status     *string  `json:"status,omitempty"`
	Rating     *int     `json:"rating,omitempty"`
	Review     *string  `json:"review,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	StartedAt  *string  `json:"started_at,omitempty"`
	FinishedAt *string  `json:"finished_at,omitempty"`
	Thumbnail  *string  `json:"thumbnail,omitempty"`
	Type       *string  `json:"type,omitempty"`
	Comments   *string  `json:"comments,omitempty"`
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
	Type       string   `dynamodbav:"type,omitempty"`
	Comments   string   `dynamodbav:"comments,omitempty"`
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
	Type       string   `json:"type,omitempty"`
	Comments   string   `json:"comments,omitempty"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

// getUserID extracts the user ID from the JWT claims in the request context
func getUserID(request events.APIGatewayProxyRequest) (string, error) {
	// For HTTP API with JWT authorizer, the claims are nested under jwt.claims
	jwt, ok := request.RequestContext.Authorizer["jwt"].(map[string]interface{})
	if !ok {
		log.Printf("Authorizer context: %+v", request.RequestContext.Authorizer)
		return "", fmt.Errorf("no jwt found in authorizer context")
	}
	
	claims, ok := jwt["claims"].(map[string]interface{})
	if !ok {
		log.Printf("JWT context: %+v", jwt)
		return "", fmt.Errorf("no claims found in jwt context")
	}
	
	// Try accessing the 'sub' claim first (standard JWT subject claim)
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	
	// Try cognito:username as fallback
	if cognitoUsername, ok := claims["cognito:username"].(string); ok {
		return cognitoUsername, nil
	}
	
	// Debug: log the actual claims structure if we can't find the user ID
	log.Printf("Claims: %+v", claims)
	
	return "", fmt.Errorf("no user ID found in JWT claims")
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from JWT claims
	userID, err := getUserID(request)
	if err != nil {
		log.Printf("Error extracting user ID: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized: Could not extract user ID",
		}, nil
	}

	// Get the book ID from path parameters
	bookID := request.PathParameters["id"]
	if bookID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Book ID is required",
		}, nil
	}

	// Parse the request body
	var updateRequest BookUpdateRequest
	if err := json.Unmarshal([]byte(request.Body), &updateRequest); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Invalid request body",
		}, nil
	}

	// Validate status if provided
	if updateRequest.Status != nil {
		validStatuses := map[string]bool{
			"WANT_TO_READ": true,
			"READING":      true,
			"READ":         true,
		}
		if !validStatuses[*updateRequest.Status] {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "Invalid status. Must be one of: WANT_TO_READ, READING, READ",
			}, nil
		}
	}

	// First, get the existing book to ensure it exists and belongs to this user
	tableNameVar := tableName
	getInput := &dynamodb.GetItemInput{
		TableName: &tableNameVar,
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "BOOK#" + bookID},
		},
	}

	getResult, err := ddbClient.GetItem(context.TODO(), getInput)
	if err != nil {
		log.Printf("Error getting item from DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	if getResult.Item == nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusNotFound,
			Body:       "Book not found",
		}, nil
	}

	// Unmarshal existing book
	var existingBook Book
	err = attributevalue.UnmarshalMap(getResult.Item, &existingBook)
	if err != nil {
		log.Printf("Error unmarshalling existing book: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Update the book with new values
	updatedBook := existingBook
	if updateRequest.Title != nil {
		updatedBook.Title = *updateRequest.Title
	}
	if updateRequest.Author != nil {
		updatedBook.Author = *updateRequest.Author
	}
	if updateRequest.Series != nil {
		updatedBook.Series = *updateRequest.Series
	}
	if updateRequest.Status != nil {
		updatedBook.Status = *updateRequest.Status
	}
	if updateRequest.Rating != nil {
		updatedBook.Rating = updateRequest.Rating
	}
	if updateRequest.Review != nil {
		updatedBook.Review = *updateRequest.Review
	}
	if updateRequest.Tags != nil {
		updatedBook.Tags = updateRequest.Tags
	}
	if updateRequest.StartedAt != nil {
		updatedBook.StartedAt = *updateRequest.StartedAt
	}
	if updateRequest.FinishedAt != nil {
		updatedBook.FinishedAt = *updateRequest.FinishedAt
	}
	if updateRequest.Thumbnail != nil {
		updatedBook.Thumbnail = *updateRequest.Thumbnail
	}
	if updateRequest.Type != nil {
		updatedBook.Type = *updateRequest.Type
	}
	if updateRequest.Comments != nil {
		updatedBook.Comments = *updateRequest.Comments
	}

	// Marshal the updated book to DynamoDB attributes
	item, err := attributevalue.MarshalMap(updatedBook)
	if err != nil {
		log.Printf("Error marshalling updated book to DynamoDB attributes: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Put the updated item in DynamoDB
	putInput := &dynamodb.PutItemInput{
		TableName: &tableNameVar,
		Item:      item,
	}

	_, err = ddbClient.PutItem(context.TODO(), putInput)
	if err != nil {
		log.Printf("Error putting updated item to DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Create the API response
	apiBook := APIBook{
		ID:         updatedBook.ID,
		Title:      updatedBook.Title,
		Author:     updatedBook.Author,
		Series:     updatedBook.Series,
		Status:     updatedBook.Status,
		Rating:     updatedBook.Rating,
		Review:     updatedBook.Review,
		Tags:       updatedBook.Tags,
		StartedAt:  updatedBook.StartedAt,
		FinishedAt: updatedBook.FinishedAt,
		Thumbnail:  updatedBook.Thumbnail,
		Type:       updatedBook.Type,
		Comments:   updatedBook.Comments,
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
		request := events.APIGatewayProxyRequest{
			PathParameters: map[string]string{
				"id": "a1b2c3d4-e5f6-7890-1234-567890abcdef", // The Way of Kings ID
			},
			Body: `{
				"rating": 8,
				"review": "Updated review after local testing",
				"status": "READ"
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
		if response.StatusCode == http.StatusOK {
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