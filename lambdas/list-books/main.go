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
	ID         string   `dynamodbav:"id"`
	PK         string   `dynamodbav:"PK"`
	SK         string   `dynamodbav:"SK"`
	Title      string   `dynamodbav:"Title"`
	Author     string   `dynamodbav:"Author"`
	Series     string   `dynamodbav:"Series"`
	Status     string   `dynamodbav:"status"`
	Rating     *int     `dynamodbav:"rating,omitempty"`
	Review     string   `dynamodbav:"review,omitempty"`
	Tags       []string `dynamodbav:"tags,omitempty"`
	StartedAt  string   `dynamodbav:"started_at,omitempty"`
	FinishedAt string   `dynamodbav:"finished_at,omitempty"`
	Thumbnail  string   `dynamodbav:"thumbnail"`
	Type       string   `dynamodbav:"type,omitempty"`
	Comments   string   `dynamodbav:"comments,omitempty"`
}

// APIBook is the structure for the API response.
type APIBook struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	Series     string   `json:"series"`
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

	status, statusOK := request.QueryStringParameters["status"]
	tableNameVar := tableName
	userPK := "USER#" + userID

	var books []Book

	if statusOK {
		// If status is provided, we need to query the user's books and filter by status
		// Since we changed the data model, we query by PK and filter by status
		queryInput := &dynamodb.QueryInput{
			TableName:              &tableNameVar,
			KeyConditionExpression: aws.String("PK = :pk"),
			FilterExpression:       aws.String("#status = :status"),
			ExpressionAttributeNames: map[string]string{
				"#status": "status",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk":     &types.AttributeValueMemberS{Value: userPK},
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
		// If no status parameter, query all books for this user
		queryInput := &dynamodb.QueryInput{
			TableName:              &tableNameVar,
			KeyConditionExpression: aws.String("PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: userPK},
			},
		}

		var queryOutput *dynamodb.QueryOutput
		queryOutput, err = ddbClient.Query(context.TODO(), queryInput)
		if err == nil {
			err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &books)
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
