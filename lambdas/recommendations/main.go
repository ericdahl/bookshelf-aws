package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "books"

var ddbClient *dynamodb.Client

// Book represents a book record from DynamoDB
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

// Recommendation represents a book recommendation
type Recommendation struct {
	Title  string `json:"title"`
	Author string `json:"author"`
	Genre  string `json:"genre"`
	Reason string `json:"reason"`
}

// RecommendationResponse is the API response structure
type RecommendationResponse struct {
	Recommendations []Recommendation `json:"recommendations"`
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
	
	if sub, ok := claims["sub"].(string); ok {
		return sub, nil
	}
	
	if cognitoUsername, ok := claims["cognito:username"].(string); ok {
		return cognitoUsername, nil
	}
	
	log.Printf("Claims: %+v", claims)
	return "", fmt.Errorf("no user ID found in JWT claims")
}

// getUserBooks fetches all books for a user from DynamoDB
func getUserBooks(userID string) ([]Book, error) {
	userPK := "USER#" + userID
	
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: userPK},
		},
	}

	queryOutput, err := ddbClient.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, fmt.Errorf("error querying DynamoDB: %v", err)
	}

	var books []Book
	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &books)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling books: %v", err)
	}

	return books, nil
}

// generateRecommendations returns mock book recommendations based on user's reading history
func generateRecommendations(books []Book) ([]Recommendation, error) {
	// For now, return mock recommendations that vary based on user's reading
	var recommendations []Recommendation
	
	if len(books) == 0 {
		// Default recommendations for users with no books
		recommendations = []Recommendation{
			{
				Title:  "The Hobbit",
				Author: "J.R.R. Tolkien",
				Genre:  "Fantasy",
				Reason: "A classic adventure perfect for starting your reading journey",
			},
			{
				Title:  "Dune",
				Author: "Frank Herbert",
				Genre:  "Science Fiction",
				Reason: "Epic world-building and complex politics",
			},
			{
				Title:  "The Name of the Wind",
				Author: "Patrick Rothfuss",
				Genre:  "Fantasy",
				Reason: "Beautiful prose and compelling storytelling",
			},
			{
				Title:  "The Martian",
				Author: "Andy Weir",
				Genre:  "Science Fiction",
				Reason: "Engaging hard sci-fi with humor",
			},
			{
				Title:  "The Way of Kings",
				Author: "Brandon Sanderson",
				Genre:  "Epic Fantasy",
				Reason: "Intricate magic system and world-building",
			},
		}
	} else {
		// Analyze user's books and provide contextual recommendations
		hasFantasy := false
		hasSciFi := false
		
		for _, book := range books {
			// Simple keyword detection to determine genres
			title := strings.ToLower(book.Title)
			tags := book.Tags
			
			for _, tag := range tags {
				switch strings.ToLower(tag) {
				case "fantasy", "epic fantasy":
					hasFantasy = true
				case "science fiction", "sci-fi", "scifi":
					hasSciFi = true
				}
			}
			
			// Also check titles for common fantasy/sci-fi indicators
			if strings.Contains(title, "dragon") || strings.Contains(title, "magic") || strings.Contains(title, "kingdom") {
				hasFantasy = true
			}
			if strings.Contains(title, "space") || strings.Contains(title, "future") || strings.Contains(title, "robot") {
				hasSciFi = true
			}
		}
		
		recommendations = []Recommendation{
			{
				Title:  "The Fifth Season",
				Author: "N.K. Jemisin",
				Genre:  "Fantasy",
				Reason: "Award-winning fantasy with unique world-building",
			},
			{
				Title:  "Project Hail Mary",
				Author: "Andy Weir",
				Genre:  "Science Fiction", 
				Reason: "Thrilling space adventure with scientific detail",
			},
			{
				Title:  "The Priory of the Orange Tree",
				Author: "Samantha Shannon",
				Genre:  "Epic Fantasy",
				Reason: "Standalone fantasy epic with dragons",
			},
			{
				Title:  "Klara and the Sun",
				Author: "Kazuo Ishiguro",
				Genre:  "Literary Fiction",
				Reason: "Thoughtful exploration of AI and humanity",
			},
			{
				Title:  "The Seven Moons of Maali Almeida",
				Author: "Shehan Karunatilaka",
				Genre:  "Magical Realism",
				Reason: "Unique perspective and engaging storytelling",
			},
		}
		
		// Adjust recommendations based on detected preferences
		if hasFantasy && !hasSciFi {
			recommendations[1] = Recommendation{
				Title:  "The Blade Itself",
				Author: "Joe Abercrombie",
				Genre:  "Grimdark Fantasy",
				Reason: "Character-driven fantasy with moral complexity",
			}
		} else if hasSciFi && !hasFantasy {
			recommendations[0] = Recommendation{
				Title:  "The Left Hand of Darkness",
				Author: "Ursula K. Le Guin",
				Genre:  "Science Fiction",
				Reason: "Groundbreaking sci-fi exploring gender and society",
			}
		}
	}
	
	return recommendations, nil
}

// handler is the Lambda function handler
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

	// Get user's books from DynamoDB
	books, err := getUserBooks(userID)
	if err != nil {
		log.Printf("Error fetching user books: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not fetch books",
		}, nil
	}

	// Generate recommendations using Bedrock
	recommendations, err := generateRecommendations(books)
	if err != nil {
		log.Printf("Error generating recommendations: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not generate recommendations",
		}, nil
	}

	// Prepare response
	response := RecommendationResponse{
		Recommendations: recommendations,
	}

	body, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling JSON response: %v", err)
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