package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const tableName = "books"

var (
	ddbClient *dynamodb.Client
	s3Client  *s3.Client
	bucketName = os.Getenv("EXPORTS_BUCKET_NAME")
)

type Book struct {
	ID         string   `dynamodbav:"id" json:"id"`
	PK         string   `dynamodbav:"PK" json:"-"`
	SK         string   `dynamodbav:"SK" json:"-"`
	Title      string   `dynamodbav:"Title" json:"title"`
	Author     string   `dynamodbav:"Author" json:"author"`
	Series     string   `dynamodbav:"Series" json:"series"`
	Status     string   `dynamodbav:"status" json:"status"`
	Rating     *int     `dynamodbav:"rating,omitempty" json:"rating,omitempty"`
	Review     string   `dynamodbav:"review,omitempty" json:"review,omitempty"`
	Tags       []string `dynamodbav:"tags,omitempty" json:"tags,omitempty"`
	StartedAt  string   `dynamodbav:"started_at,omitempty" json:"started_at,omitempty"`
	FinishedAt string   `dynamodbav:"finished_at,omitempty" json:"finished_at,omitempty"`
	Thumbnail  string   `dynamodbav:"thumbnail" json:"thumbnail"`
	Type       string   `dynamodbav:"type,omitempty" json:"type,omitempty"`
	Comments   string   `dynamodbav:"comments,omitempty" json:"comments,omitempty"`
}

type ExportRequest struct {
	Format  string            `json:"format"`
	Filters map[string]string `json:"filters,omitempty"`
}

type ExportResponse struct {
	DownloadURL string `json:"download_url"`
	Format      string `json:"format"`
	Filename    string `json:"filename"`
	ExpiresAt   string `json:"expires_at"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
	s3Client = s3.NewFromConfig(cfg)
}

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

func getAllUserBooks(ctx context.Context, userID string, filters map[string]string) ([]Book, error) {
	userPK := "USER#" + userID
	
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: userPK},
		},
	}

	// Apply status filter if provided
	if status, exists := filters["status"]; exists {
		queryInput.FilterExpression = aws.String("#status = :status")
		queryInput.ExpressionAttributeNames = map[string]string{
			"#status": "status",
		}
		queryInput.ExpressionAttributeValues[":status"] = &types.AttributeValueMemberS{Value: status}
	}

	result, err := ddbClient.Query(ctx, queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query books: %v", err)
	}

	var books []Book
	err = attributevalue.UnmarshalListOfMaps(result.Items, &books)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal books: %v", err)
	}

	return books, nil
}

func generateCSV(books []Book) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{
		"Title", "Author", "Series", "Status", "Rating", 
		"Started Date", "Finished Date", "Tags", "Type", 
		"Review", "Comments", "Thumbnail",
	}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	// Write data rows
	for _, book := range books {
		rating := ""
		if book.Rating != nil {
			rating = strconv.Itoa(*book.Rating)
		}
		
		tags := strings.Join(book.Tags, "; ")
		
		record := []string{
			book.Title,
			book.Author,
			book.Series,
			book.Status,
			rating,
			book.StartedAt,
			book.FinishedAt,
			tags,
			book.Type,
			book.Review,
			book.Comments,
			book.Thumbnail,
		}
		
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

func generateJSON(books []Book) ([]byte, error) {
	return json.MarshalIndent(books, "", "  ")
}

func uploadToS3(ctx context.Context, data []byte, filename string) (string, error) {
	if bucketName == "" {
		return "", fmt.Errorf("EXPORTS_BUCKET_NAME environment variable not set")
	}

	// Upload file to S3
	_, err := s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
		Metadata: map[string]string{
			"created-at": time.Now().UTC().Format(time.RFC3339),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// Generate pre-signed URL (15 minutes expiry)
	presigner := s3.NewPresignClient(s3Client)
	request, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(filename),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = time.Minute * 15
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate pre-signed URL: %v", err)
	}

	return request.URL, nil
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract user ID from JWT claims
	userID, err := getUserID(request)
	if err != nil {
		log.Printf("Error extracting user ID: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Unauthorized: Could not extract user ID"}`,
		}, nil
	}

	// Parse request body
	var exportReq ExportRequest
	if request.Body != "" {
		if err := json.Unmarshal([]byte(request.Body), &exportReq); err != nil {
			return events.APIGatewayProxyResponse{
				StatusCode: http.StatusBadRequest,
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: `{"error": "Invalid JSON in request body"}`,
			}, nil
		}
	}

	// Default format to CSV if not specified
	if exportReq.Format == "" {
		exportReq.Format = "csv"
	}

	// Validate format
	if exportReq.Format != "csv" && exportReq.Format != "json" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Invalid format. Supported formats: csv, json"}`,
		}, nil
	}

	// Get all user's books
	books, err := getAllUserBooks(ctx, userID, exportReq.Filters)
	if err != nil {
		log.Printf("Error getting user books: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to retrieve books"}`,
		}, nil
	}

	// Generate export data
	var data []byte
	var fileExtension string
	
	switch exportReq.Format {
	case "csv":
		data, err = generateCSV(books)
		fileExtension = "csv"
	case "json":
		data, err = generateJSON(books)
		fileExtension = "json"
	}

	if err != nil {
		log.Printf("Error generating export data: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to generate export data"}`,
		}, nil
	}

	// Generate unique filename
	timestamp := time.Now().UTC().Format("20060102-150405")
	filename := fmt.Sprintf("exports/%s/books-%s.%s", userID, timestamp, fileExtension)

	// Upload to S3 and get signed URL
	downloadURL, err := uploadToS3(ctx, data, filename)
	if err != nil {
		log.Printf("Error uploading to S3: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to upload export file"}`,
		}, nil
	}

	// Prepare response
	response := ExportResponse{
		DownloadURL: downloadURL,
		Format:      exportReq.Format,
		Filename:    fmt.Sprintf("books-%s.%s", timestamp, fileExtension),
		ExpiresAt:   time.Now().Add(15 * time.Minute).UTC().Format(time.RFC3339),
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Failed to generate response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	if os.Getenv("LAMBDA_TASK_ROOT") == "" {
		// Local testing
		fmt.Println("--- Local execution mode ---")
		fmt.Println("Set EXPORTS_BUCKET_NAME environment variable for S3 operations")
		
		// Create a test request
		testReq := events.APIGatewayProxyRequest{
			Body: `{"format": "csv"}`,
			RequestContext: events.APIGatewayProxyRequestContext{
				Authorizer: map[string]interface{}{
					"jwt": map[string]interface{}{
						"claims": map[string]interface{}{
							"sub": "test-user-id",
						},
					},
				},
			},
		}
		
		response, err := handler(context.Background(), testReq)
		if err != nil {
			log.Fatalf("FATAL: handler failed: %v", err)
		}
		
		fmt.Printf("Status Code: %d\n", response.StatusCode)
		fmt.Printf("Response Body: %s\n", response.Body)
	} else {
		lambda.Start(handler)
	}
}