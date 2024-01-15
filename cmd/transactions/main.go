package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"transactions/internal/db"
)

const (
	INSERT_TIMEOUT = 10 * time.Second
	LIST_TIMEOUT   = 5 * time.Second
)

var client *db.Client

func init() {
	client = db.NewClient()
}

func handleError(format string, args ...interface{}) (events.APIGatewayProxyResponse, error) {
	err := fmt.Errorf(format, args...)
	return events.APIGatewayProxyResponse{
		Body:       err.Error(),
		StatusCode: http.StatusInternalServerError,
	}, nil
}

func handleOK(body interface{}) (events.APIGatewayProxyResponse, error) {
	json, err := json.Marshal(body)
	if err != nil {
		return handleError("failed to convert response body to JSON: %w", err)
	}
	return events.APIGatewayProxyResponse{
		Body:       string(json),
		StatusCode: http.StatusOK,
	}, nil
}

func handleCreate(
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	var tr db.Transaction

	dec := json.NewDecoder(strings.NewReader(request.Body))
	if err := dec.Decode(&tr); err != nil {
		return handleError("failed to decode request body: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), INSERT_TIMEOUT)
	defer cancel()
	if err := client.Create(ctx, &tr); err != nil {
		return handleError("failed to create record: %w", err)
	}

	return handleOK(tr)
}

func handleList(
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	log.Printf("request path params: %v", req.PathParameters)
	ctx, cancel := context.WithTimeout(context.Background(), LIST_TIMEOUT)
	defer cancel()

	listReq, err := db.UserListRequestFromAPIGatewayProxyRequest(req)
	if err != nil {
		return handleError("failed to parse request: %w", err)
	}

	trs, err := client.Query(ctx, listReq)
	if err != nil {
		return handleError("failed to scan records in dynamodb: %w", err)
	}

	return handleOK(trs)
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
