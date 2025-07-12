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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "books"

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

// Book represents a book record from DynamoDB.
type Book struct {
	ID        string `dynamodbav:"id"`
	PK        string `dynamodbav:"PK"`
	SK        string `dynamodbav:"SK"`
	Title     string `dynamodbav:"Title"`
	Author    string `dynamodbav:"Author"`
	Series    string `dynamodbav:"Series"`
	Status    string `dynamodbav:"status"`
	Thumbnail string `dynamodbav:"thumbnail"`
}

// APIBook is the structure for the API response.
type APIBook struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Series    string `json:"series"`
	Status    string `json:"status"`
	Thumbnail string `json:"thumbnail,omitempty"`
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
	status, statusOK := request.QueryStringParameters["status"]
	tableNameVar := tableName

	var books []Book
	var err error

	if statusOK {
		// If status is provided, query the GSI
		gsiName := "status-gsi"
		queryInput := &dynamodb.QueryInput{
			TableName:              &tableNameVar,
			IndexName:              &gsiName,
			KeyConditionExpression: aws.String("#status = :status"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":status": &types.AttributeValueMemberS{Value: status},
			},
		}

		var queryOutput *dynamodb.QueryOutput
		queryOutput, err = ddbClient.Query(context.TODO(), queryInput)
		if err != nil {
			log.Printf("Error querying DynamoDB: %v", err)
		}
		if err == nil {
			log.Printf("DynamoDB items: %v", queryOutput.Items)
			err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &books)
			if err != nil {
				log.Printf("Error unmarshalling items: %v", err)
			}
		}
	} else {
		// If no status parameter, scan the table
		var scanOutput *dynamodb.ScanOutput
		scanOutput, err = ddbClient.Scan(context.TODO(), &dynamodb.ScanInput{
			TableName: &tableNameVar,
		})
		if err == nil {
			err = attributevalue.UnmarshalListOfMaps(scanOutput.Items, &books)
		}
	}

	if err != nil {
		log.Printf("Error processing DynamoDB request: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not process book data",
		}, nil
	}

	if statusOK && len(books) == 0 {
		body, _ := json.Marshal([]APIBook{})
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusOK,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: string(body),
		}, nil
	}

	// Map to API response structure
	apiBooks := make([]APIBook, len(books))
	for i, book := range books {
		apiBooks[i] = APIBook{
			ID:        book.ID,
			Title:     book.Title,
			Author:    book.Author,
			Series:    book.Series,
			Status:    book.Status,
			Thumbnail: book.Thumbnail,
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
