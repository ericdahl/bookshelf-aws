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
)

const tableName = "books"

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

// Book represents a book record from DynamoDB.
type Book struct {
	ID     string `dynamodbav:"id"`
	PK     string `dynamodbav:"PK"`
	Title  string `dynamodbav:"Title"`
	Author string `dynamodbav:"Author"`
	Series string `dynamodbav:"Series"`
	Status string `dynamodbav:"status"`
}

// APIBook is the structure for the API response.
type APIBook struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Series string `json:"series"`
	Status string `json:"status"`
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
	// For this version, we scan the table to get all books.
	// For a production application with a large dataset, a more efficient
	// query approach would be recommended over a full table scan.
	tableNameVar := tableName
	scanOutput, err := ddbClient.Scan(context.TODO(), &dynamodb.ScanInput{
		TableName: &tableNameVar,
	})
	if err != nil {
		log.Printf("Error scanning DynamoDB: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not scan books",
		}, nil
	}

	var books []Book
	err = attributevalue.UnmarshalListOfMaps(scanOutput.Items, &books)
	if err != nil {
		log.Printf("Error unmarshalling DynamoDB items: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not process book data",
		}, nil
	}

	// Map to API response structure
	apiBooks := make([]APIBook, len(books))
	for i, book := range books {
		apiBooks[i] = APIBook{
			ID:     book.ID,
			Title:  book.Title,
			Author: book.Author,
			Series: book.Series,
			Status: book.Status,
		}
	}

	body, err := json.Marshal(apiBooks)
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