package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

const tableName = "books"

// Cognito configuration
var (
	cognitoUserPoolID = getRequiredEnv("COGNITO_USER_POOL_ID")
	cognitoRegion     = getRequiredEnv("COGNITO_REGION")
)

// getRequiredEnv gets environment variable value or panics if not set
func getRequiredEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Fatalf("Required environment variable %s is not set", key)
	return "" // Never reached, but needed for compilation
}

// getEnv gets environment variable value or returns default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ddbClient is the DynamoDB client.
var ddbClient *dynamodb.Client

// jwksCache stores the JWKS keys from Cognito
var jwksCache map[string]*rsa.PublicKey
var jwksLastFetch time.Time

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
	Thumbnail string `json:"thumbnail"`
}

// JWKSResponse represents the JWKS response structure
type JWKSResponse struct {
	Keys []JWKSKey `json:"keys"`
}

// JWKSKey represents a single key in the JWKS response
type JWKSKey struct {
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Kty string `json:"kty"`
	Use string `json:"use"`
}

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	ddbClient = dynamodb.NewFromConfig(cfg)
	jwksCache = make(map[string]*rsa.PublicKey)
}

// validateJWTToken validates the JWT token from the Authorization header
func validateJWTToken(authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("missing Authorization header")
	}

	// Extract Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return fmt.Errorf("invalid Authorization header format")
	}

	tokenString := parts[1]

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from the token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("missing kid in token header")
		}

		// Get the public key for this kid
		publicKey, err := getPublicKey(kid)
		if err != nil {
			return nil, fmt.Errorf("failed to get public key: %v", err)
		}

		return publicKey, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %v", err)
	}

	if !token.Valid {
		return fmt.Errorf("invalid token")
	}

	// Validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("invalid claims")
	}

	// Check token_use claim
	tokenUse, ok := claims["token_use"].(string)
	if !ok || tokenUse != "access" {
		return fmt.Errorf("invalid token_use claim")
	}

	// Check issuer
	iss, ok := claims["iss"].(string)
	if !ok {
		return fmt.Errorf("missing iss claim")
	}

	expectedIss := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", cognitoRegion, cognitoUserPoolID)
	if iss != expectedIss {
		return fmt.Errorf("invalid issuer")
	}

	// Check expiration
	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("missing exp claim")
	}

	if time.Now().Unix() > int64(exp) {
		return fmt.Errorf("token expired")
	}

	return nil
}

// getPublicKey retrieves the public key for the given kid
func getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check cache first
	if publicKey, exists := jwksCache[kid]; exists {
		// Check if cache is still valid (cache for 1 hour)
		if time.Since(jwksLastFetch) < time.Hour {
			return publicKey, nil
		}
	}

	// Fetch JWKS from Cognito
	jwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", cognitoRegion, cognitoUserPoolID)
	
	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	var jwksResponse JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwksResponse); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS response: %v", err)
	}

	// Find the key with matching kid
	for _, key := range jwksResponse.Keys {
		if key.Kid == kid {
			publicKey, err := convertJWKSKeyToRSA(key)
			if err != nil {
				return nil, fmt.Errorf("failed to convert JWKS key to RSA: %v", err)
			}

			// Update cache
			jwksCache[kid] = publicKey
			jwksLastFetch = time.Now()

			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("key not found for kid: %s", kid)
}

// convertJWKSKeyToRSA converts a JWKS key to RSA public key
func convertJWKSKeyToRSA(key JWKSKey) (*rsa.PublicKey, error) {
	// Create a JWK from the key data - need to convert to JSON first
	jwkData := map[string]interface{}{
		"kty": key.Kty,
		"use": key.Use,
		"kid": key.Kid,
		"n":   key.N,
		"e":   key.E,
	}
	
	// Convert to JSON bytes
	jwkBytes, err := json.Marshal(jwkData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JWK data: %v", err)
	}
	
	// Parse JWK from JSON
	jwkKey, err := jwk.ParseKey(jwkBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWK: %v", err)
	}

	// Convert to RSA public key
	var rsaKey rsa.PublicKey
	if err := jwkKey.Raw(&rsaKey); err != nil {
		return nil, fmt.Errorf("failed to convert JWK to RSA: %v", err)
	}

	return &rsaKey, nil
}

// handler is the Lambda function handler.
func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Validate JWT token
	authHeader := request.Headers["Authorization"]
	if authHeader == "" {
		authHeader = request.Headers["authorization"] // case-insensitive header lookup
	}

	if err := validateJWTToken(authHeader); err != nil {
		log.Printf("Authentication failed: %v", err)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusUnauthorized,
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: `{"error": "Unauthorized"}`,
		}, nil
	}

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
