package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/google/uuid"
)

type Transaction struct {
	ID            string  `json:"id,omitempty"        danamodbav:"ID"`
	Timestamp     int64   `json:"timestamp,omitempty"                 dynamodbav:"Timestamp"`
	UserID        string  `json:"user_id"                             dynamodbav:"UserID"`
	Origin        string  `json:"origin"                              dynamodbav:"Origin"`
	OperationType string  `json:"operation_type"                      dynamodbav:"OperationType"`
	Amount        float64 `json:"amount"                              dynamodbav:"Amount"`
}

// Convert tr to DynamoDB AttributeValue map
func (tr *Transaction) SetDefaults() {
	if tr.ID == "" {
		tr.ID = uuid.New().String()
	}

	if tr.Timestamp == 0 {
		tr.Timestamp = time.Now().Unix()
	}
}

const (
	INSERT_TIMEOUT = 1 * time.Second
	LIST_TIMEOUT   = 2 * time.Second
)

var (
	dynamodbClient *dynamodb.Client
	tableName      string
)

func ConnectDynamoDB() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), func(opts *config.LoadOptions) error {
		// opts.Region = "us-east-1"
		return nil
	})
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

func init() {
	dynamodbClient = ConnectDynamoDB()
	tableName, _ = os.LookupEnv("TABLE_NAME")
}

func handleError(format string, args ...interface{}) (events.APIGatewayProxyResponse, error) {
	err := fmt.Errorf(format, args...)
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: http.StatusInternalServerError,
	}, nil
}

func handleCreate(
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	var tr Transaction

	dec := json.NewDecoder(strings.NewReader(request.Body))
	if err := dec.Decode(&tr); err != nil {
		return handleError("failed to decode request body: %w", err)
	}

	tr.SetDefaults()

	// Convert tr to DynamoDB AttributeValue map
	trAv, err := attributevalue.MarshalMap(tr)
	if err != nil {
		return handleError("failed to convert transaction to dynamodb attribute value: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), INSERT_TIMEOUT)
	defer cancel()
	if _, err = dynamodbClient.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      trAv,
	}); err != nil {
		return handleError("failed to insert record into dynamodb: %w", err)
	}

	return events.APIGatewayProxyResponse{
		Body:       "OK",
		StatusCode: http.StatusOK,
	}, nil
}

func handleList(
	events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), LIST_TIMEOUT)
	defer cancel()

	result, err := dynamodbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return handleError("failed to scan records in dynamodb: %w", err)
	}

	var trs []Transaction

	// Using Go SDK helper function to parse dynamodb attributes to struct
	if err = attributevalue.UnmarshalListOfMaps(result.Items, &trs); err != nil {
		return handleError("failed to convert dynamodb scan result into transaction list: %w", err)
	}

	json, err := json.Marshal(trs)
	if err != nil {
		return handleError("failed to convert transaction list to JSON: %w", err)
	}

	return events.APIGatewayProxyResponse{
		Body:       string(json),
		StatusCode: http.StatusOK,
	}, nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch request.HTTPMethod {
	case "POST":
		return handleCreate(request)
	case "GET":
		return handleList(request)
	default:
		return events.APIGatewayProxyResponse{
			Body:       "Unsupported method",
			StatusCode: http.StatusBadRequest,
		}, nil
	}
}

func main() {
	lambda.Start(handler)
}
