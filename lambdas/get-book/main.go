package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Series    string `json:"series"`
	Status    string `json:"status"`
	Thumbnail string `json:"thumbnail"`
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id := request.PathParameters["id"]
	if id == "" {
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "book ID is required"}, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("failed to load configuration, %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	client := dynamodb.NewFromConfig(cfg)

	input := &dynamodb.GetItemInput{
		TableName: &[]string{"books"}[0],
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "BOOK#" + id},
			"SK": &types.AttributeValueMemberS{Value: "BOOK"},
		},
	}

	result, err := client.GetItem(ctx, input)
	if err != nil {
		log.Printf("failed to get item, %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	if result.Item == nil {
		return events.APIGatewayProxyResponse{StatusCode: 404, Body: "book not found"}, nil
	}

	book := Book{
		ID:     result.Item["id"].(*types.AttributeValueMemberS).Value,
		Title:  result.Item["Title"].(*types.AttributeValueMemberS).Value,
		Author: result.Item["Author"].(*types.AttributeValueMemberS).Value,
		Series: result.Item["Series"].(*types.AttributeValueMemberS).Value,
		Status: result.Item["status"].(*types.AttributeValueMemberS).Value,
	}
	
	// Handle optional thumbnail field
	if thumbnailAttr, exists := result.Item["thumbnail"]; exists {
		if thumbnailValue, ok := thumbnailAttr.(*types.AttributeValueMemberS); ok {
			book.Thumbnail = thumbnailValue.Value
		}
	}

	body, err := json.Marshal(book)
	if err != nil {
		log.Printf("failed to marshal book, %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
		Body:       string(body),
	}, nil
}

func main() {
	// If the LAMBDA_TASK_ROOT environment variable is not set, we're running locally.
	if os.Getenv("LAMBDA_TASK_ROOT") == "" {
		// Create a dummy request for local testing with a sample book ID.
		request := events.APIGatewayProxyRequest{
			PathParameters: map[string]string{
				"id": "a1b2c3d4-e5f6-7890-1234-567890abcdef", // The Way of Kings ID
			},
		}

		// Call the handler directly.
		response, err := HandleRequest(context.Background(), request)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}

		// Print the response details to stdout.
		fmt.Println("--- Local execution ---")
		fmt.Printf("Status Code: %d\n", response.StatusCode)
		if response.StatusCode == 200 {
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
		lambda.Start(HandleRequest)
	}
}
