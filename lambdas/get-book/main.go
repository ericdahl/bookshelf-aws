package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Book struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
	Series string `json:"series"`
	Status string `json:"status"`
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
	lambda.Start(HandleRequest)
}
