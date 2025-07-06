package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "books"

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get the book ID from path parameters
	bookID := request.PathParameters["id"]
	if bookID == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Body:       "Book ID is required",
		}, nil
	}

	// First, check if the book exists
	tableNameVar := tableName
	getInput := &dynamodb.GetItemInput{
		TableName: &tableNameVar,
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "BOOK#" + bookID},
			"SK": &types.AttributeValueMemberS{Value: "BOOK"},
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

	// Delete the book from DynamoDB
	deleteInput := &dynamodb.DeleteItemInput{
		TableName: &tableNameVar,
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "BOOK#" + bookID},
			"SK": &types.AttributeValueMemberS{Value: "BOOK"},
		},
	}

	_, err = ddbClient.DeleteItem(context.TODO(), deleteInput)
	if err != nil {
		log.Printf("Error deleting item from DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	// Return 204 No Content on successful deletion
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNoContent,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: "",
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
		}

		// Call the handler directly.
		response, err := handler(request)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}

		// Print the response details to stdout.
		fmt.Println("--- Local execution ---")
		fmt.Printf("Status Code: %d\n", response.StatusCode)
		fmt.Printf("Response Body: %s\n", response.Body)
	} else {
		// Start the Lambda handler in the AWS environment.
		lambda.Start(handler)
	}
} 