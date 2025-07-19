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

// getUserID extracts the user ID from the JWT claims in the request context
func getUserID(request events.APIGatewayProxyRequest) (string, error) {
	// JWT claims are available in the request context when using API Gateway JWT authorizer
	// Try different possible structures
	
	// Try accessing claims directly under authorizer
	if sub, ok := request.RequestContext.Authorizer["sub"].(string); ok {
		return sub, nil
	}
	
	// Try accessing under jwt key
	if jwt, ok := request.RequestContext.Authorizer["jwt"].(map[string]interface{}); ok {
		if claims, ok := jwt["claims"].(map[string]interface{}); ok {
			if sub, ok := claims["sub"].(string); ok {
				return sub, nil
			}
		}
		// Try direct access from jwt
		if sub, ok := jwt["sub"].(string); ok {
			return sub, nil
		}
	}
	
	// Debug: log the actual structure
	log.Printf("Authorizer context: %+v", request.RequestContext.Authorizer)
	
	return "", fmt.Errorf("no sub claim found in JWT context")
}

func HandleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from JWT claims
	userID, err := getUserID(request)
	if err != nil {
		log.Printf("Error extracting user ID: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized: Could not extract user ID",
		}, nil
	}

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
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "BOOK#" + id},
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

	var book Book
	err = attributevalue.UnmarshalMap(result.Item, &book)
	if err != nil {
		log.Printf("Error unmarshalling book: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Internal Server Error"}, nil
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
		Type:       book.Type,
		Comments:   book.Comments,
	}

	body, err := json.Marshal(apiBook)
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
