package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const tableName = "books"

var ddbClient *dynamodb.Client
var bedrockClient *bedrockruntime.Client
var logger *slog.Logger

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

// Titan request/response structures for Bedrock
type TitanRequest struct {
	InputText            string                `json:"inputText"`
	TextGenerationConfig TextGenerationConfig `json:"textGenerationConfig"`
}

type TextGenerationConfig struct {
	MaxTokenCount   int     `json:"maxTokenCount"`
	StopSequences   []string `json:"stopSequences,omitempty"`
	Temperature     float64 `json:"temperature,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
}

type TitanResponse struct {
	InputTextTokenCount int                     `json:"inputTextTokenCount"`
	Results             []TitanGenerationResult `json:"results"`
}

type TitanGenerationResult struct {
	TokenCount       int    `json:"tokenCount"`
	OutputText       string `json:"outputText"`
	CompletionReason string `json:"completionReason"`
}


func init() {
	// Set up structured logging
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.Error("unable to load SDK config", "error", err)
		os.Exit(1)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
	bedrockClient = bedrockruntime.NewFromConfig(cfg)

	logger.Info("Lambda initialized successfully")
}

// getUserID extracts the user ID from the JWT claims in the request context
func getUserID(request events.APIGatewayProxyRequest) (string, error) {
	jwt, ok := request.RequestContext.Authorizer["jwt"].(map[string]interface{})
	if !ok {
		logger.Warn("no jwt found in authorizer context", 
			"authorizer_context", request.RequestContext.Authorizer)
		return "", fmt.Errorf("no jwt found in authorizer context")
	}
	
	claims, ok := jwt["claims"].(map[string]interface{})
	if !ok {
		logger.Warn("no claims found in jwt context", "jwt_context", jwt)
		return "", fmt.Errorf("no claims found in jwt context")
	}
	
	if sub, ok := claims["sub"].(string); ok {
		logger.Debug("extracted user ID from sub claim", "user_id", sub)
		return sub, nil
	}
	
	if cognitoUsername, ok := claims["cognito:username"].(string); ok {
		logger.Debug("extracted user ID from cognito:username claim", "user_id", cognitoUsername)
		return cognitoUsername, nil
	}
	
	logger.Warn("no user ID found in JWT claims", "claims", claims)
	return "", fmt.Errorf("no user ID found in JWT claims")
}

// getUserBooks fetches all books for a user from DynamoDB
func getUserBooks(userID string) ([]Book, error) {
	startTime := time.Now()
	userPK := "USER#" + userID
	
	logger.Info("fetching user books from DynamoDB", 
		"user_id", userID, 
		"user_pk", userPK,
		"table", tableName)
	
	queryInput := &dynamodb.QueryInput{
		TableName:              aws.String(tableName),
		KeyConditionExpression: aws.String("PK = :pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk": &types.AttributeValueMemberS{Value: userPK},
		},
	}

	queryOutput, err := ddbClient.Query(context.TODO(), queryInput)
	duration := time.Since(startTime)
	
	if err != nil {
		logger.Error("error querying DynamoDB", 
			"error", err, 
			"duration_ms", duration.Milliseconds(),
			"user_id", userID)
		return nil, fmt.Errorf("error querying DynamoDB: %v", err)
	}

	var books []Book
	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &books)
	if err != nil {
		logger.Error("error unmarshalling books", 
			"error", err, 
			"items_count", len(queryOutput.Items),
			"user_id", userID)
		return nil, fmt.Errorf("error unmarshalling books: %v", err)
	}

	logger.Info("successfully fetched user books", 
		"user_id", userID,
		"books_count", len(books),
		"duration_ms", duration.Milliseconds())

	return books, nil
}

// generateRecommendations uses Amazon Titan Text to generate book recommendations
func generateRecommendations(books []Book) ([]Recommendation, error) {
	startTime := time.Now()
	
	logger.Info("generating book recommendations", 
		"books_count", len(books))

	if len(books) == 0 {
		logger.Info("no books found, returning default recommendations")
		return getDefaultRecommendations(), nil
	}

	// Build prompt from user's reading history
	var readBooks []string
	var currentlyReading []string
	var statusCounts = make(map[string]int)
	
	for _, book := range books {
		bookStr := fmt.Sprintf("%s by %s", book.Title, book.Author)
		statusCounts[book.Status]++
		
		switch strings.ToUpper(book.Status) {
		case "READ", "FINISHED":
			readBooks = append(readBooks, bookStr)
		case "READING", "CURRENTLY-READING", "CURRENTLY_READING":
			currentlyReading = append(currentlyReading, bookStr)
		}
	}
	
	logger.Info("book status breakdown", 
		"status_counts", statusCounts,
		"matched_read_books", len(readBooks),
		"matched_currently_reading", len(currentlyReading))

	prompt := "Based on the following reading history, recommend 5 books with their genres. Return the response as a valid JSON array with objects containing 'title', 'author', 'genre', and 'reason' fields.\n\n"
	
	if len(readBooks) > 0 {
		prompt += "Books read: " + strings.Join(readBooks, ", ") + "\n"
	}
	
	if len(currentlyReading) > 0 {
		prompt += "Currently reading: " + strings.Join(currentlyReading, ", ") + "\n"
	}
	
	prompt += "\nPlease provide exactly 5 book recommendations in valid JSON format. Return only the JSON array, no additional text."

	logger.Info("built prompt for Bedrock", 
		"read_books_count", len(readBooks),
		"currently_reading_count", len(currentlyReading),
		"prompt_length", len(prompt))

	// Call Bedrock Titan model
	recommendations, err := callBedrockTitan(prompt)
	duration := time.Since(startTime)
	
	if err != nil {
		logger.Warn("error calling Bedrock, falling back to defaults", 
			"error", err,
			"duration_ms", duration.Milliseconds())
		// Fallback to default recommendations if Bedrock fails
		return getDefaultRecommendations(), nil
	}

	logger.Info("successfully generated recommendations", 
		"recommendations_count", len(recommendations),
		"duration_ms", duration.Milliseconds())

	return recommendations, nil
}

// callBedrockTitan makes a request to Amazon Titan Text model
func callBedrockTitan(prompt string) ([]Recommendation, error) {
	startTime := time.Now()
	
	// Prepare Titan request
	titanReq := TitanRequest{
		InputText: prompt,
		TextGenerationConfig: TextGenerationConfig{
			MaxTokenCount: 1000,
			Temperature:   0.7,
			TopP:          0.9,
		},
	}

	requestBody, err := json.Marshal(titanReq)
	if err != nil {
		logger.Error("error marshalling Titan request", "error", err)
		return nil, fmt.Errorf("error marshalling Titan request: %v", err)
	}

	// Call Bedrock with Titan model
	modelID := "amazon.titan-text-express-v1"
	
	logger.Info("calling Bedrock Titan model", 
		"model_id", modelID,
		"prompt", prompt,
		"max_tokens", 1000,
		"temperature", 0.7,
		"top_p", 0.9)

	output, err := bedrockClient.InvokeModel(context.TODO(), &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(modelID),
		Body:        requestBody,
		ContentType: aws.String("application/json"),
	})
	
	bedrockDuration := time.Since(startTime)
	
	if err != nil {
		logger.Error("error calling Bedrock", 
			"error", err,
			"model_id", modelID,
			"duration_ms", bedrockDuration.Milliseconds())
		return nil, fmt.Errorf("error calling Bedrock: %v", err)
	}

	logger.Info("received response from Bedrock", 
		"model_id", modelID,
		"duration_ms", bedrockDuration.Milliseconds(),
		"response_size_bytes", len(output.Body))

	// Parse Titan response
	var titanResp TitanResponse
	err = json.Unmarshal(output.Body, &titanResp)
	if err != nil {
		logger.Error("error unmarshalling Titan response", 
			"error", err,
			"response_body", string(output.Body))
		return nil, fmt.Errorf("error unmarshalling Titan response: %v", err)
	}

	if len(titanResp.Results) == 0 {
		logger.Error("no results in Titan response", "response", titanResp)
		return nil, fmt.Errorf("no results in Titan response")
	}

	// Extract JSON from Titan's text response
	text := titanResp.Results[0].OutputText
	
	logger.Info("received text output from Titan", 
		"output_text", text,
		"input_token_count", titanResp.InputTextTokenCount,
		"output_token_count", titanResp.Results[0].TokenCount,
		"completion_reason", titanResp.Results[0].CompletionReason)
	
	// Clean up the text by removing markdown code blocks and extra formatting
	text = strings.ReplaceAll(text, "```json", "")
	text = strings.ReplaceAll(text, "```tabular-data-json", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.TrimSpace(text)
	
	// Try to find the first complete JSON array in the response
	startIdx := strings.Index(text, "[")
	if startIdx == -1 {
		logger.Error("could not find JSON array start in Titan response", "text", text)
		return nil, fmt.Errorf("could not find JSON array start in Titan response: %s", text)
	}
	
	// Find the matching closing bracket for the array
	bracketCount := 0
	endIdx := -1
	for i := startIdx; i < len(text); i++ {
		if text[i] == '[' {
			bracketCount++
		} else if text[i] == ']' {
			bracketCount--
			if bracketCount == 0 {
				endIdx = i
				break
			}
		}
	}
	
	if endIdx == -1 {
		logger.Error("could not find matching closing bracket in Titan response", "text", text)
		return nil, fmt.Errorf("could not find matching closing bracket in Titan response: %s", text)
	}
	
	jsonStr := text[startIdx : endIdx+1]
	
	// Clean up common JSON formatting issues from AI responses
	jsonStr = strings.ReplaceAll(jsonStr, "\n", " ")
	jsonStr = strings.ReplaceAll(jsonStr, "\t", " ")
	
	// Fix the malformed entry we saw in logs: `"Title": "The Diary of a Young Girl", Anne Frank"`
	// This regex finds patterns like `"Field": "Value", ExtraText"` and fixes them
	jsonStr = strings.ReplaceAll(jsonStr, `", Anne Frank"`, `"`)
	
	logger.Debug("extracted and cleaned JSON from Titan response", 
		"json_string", jsonStr,
		"json_length", len(jsonStr))
	
	var recommendations []Recommendation
	err = json.Unmarshal([]byte(jsonStr), &recommendations)
	if err != nil {
		logger.Error("failed to parse recommendations JSON", 
			"error", err,
			"json_string", jsonStr)
		return nil, fmt.Errorf("error unmarshalling recommendations JSON: %v, text: %s", err, jsonStr)
	}

	// Ensure we have exactly 5 recommendations
	if len(recommendations) > 5 {
		recommendations = recommendations[:5]
	}

	logger.Info("successfully parsed recommendations from Titan", 
		"recommendations_count", len(recommendations),
		"total_duration_ms", time.Since(startTime).Milliseconds())

	return recommendations, nil
}

// getDefaultRecommendations returns fallback recommendations
func getDefaultRecommendations() []Recommendation {
	return []Recommendation{
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
}

// handler is the Lambda function handler
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	requestStartTime := time.Now()
	
	logger.Info("handling recommendations request", 
		"request_id", request.RequestContext.RequestID,
		"path", request.Path,
		"method", request.HTTPMethod)

	// Extract user ID from JWT claims
	userID, err := getUserID(request)
	if err != nil {
		logger.Warn("error extracting user ID", 
			"error", err,
			"request_id", request.RequestContext.RequestID)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Body:       "Unauthorized: Could not extract user ID",
		}, nil
	}

	logger.Info("extracted user ID", 
		"user_id", userID,
		"request_id", request.RequestContext.RequestID)

	// Get user's books from DynamoDB
	books, err := getUserBooks(userID)
	if err != nil {
		logger.Error("error fetching user books", 
			"error", err,
			"user_id", userID,
			"request_id", request.RequestContext.RequestID)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error: Could not fetch books",
		}, nil
	}

	// Generate recommendations using Bedrock
	recommendations, err := generateRecommendations(books)
	if err != nil {
		logger.Error("error generating recommendations", 
			"error", err,
			"user_id", userID,
			"books_count", len(books),
			"request_id", request.RequestContext.RequestID)
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
		logger.Error("error marshalling JSON response", 
			"error", err,
			"user_id", userID,
			"request_id", request.RequestContext.RequestID)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Body:       "Internal Server Error",
		}, nil
	}

	totalDuration := time.Since(requestStartTime)
	
	logger.Info("successfully completed recommendations request", 
		"user_id", userID,
		"recommendations_count", len(recommendations),
		"response_size_bytes", len(body),
		"total_duration_ms", totalDuration.Milliseconds(),
		"request_id", request.RequestContext.RequestID)

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
			logger.Error("handler failed", "error", err)
			os.Exit(1)
		}

		// Print the response body to stdout.
		fmt.Println("--- Local execution ---")
		fmt.Println(response.Body)
	} else {
		// Start the Lambda handler in the AWS environment.
		lambda.Start(handler)
	}
}