package db

import (
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/go-playground/validator/v10"
)

// ListResponse represents a list of transactions with an optional cursor for the next page.
type ListResponse struct {
	Items  []Transaction `json:"items"`
	Cursor string        `json:"cursor,omitempty"`
}

// TransactionListRequest represents a query to list transactions.
type UserListRequest struct {
	UserID          string `validate:"required"` // partition key
	TimestampPrefix string `validate:"required"` // sort key to use as a prefix. Examples: "2020-01", "2020-01-01"
	Origin          string // filter by origin
	OperationType   string // filter by operation type
	After           string // cursor for the next page
	Limit           *int32 // max number of items to return
}

// UserListRequestFromAPIGatewayProxyRequest converts an API Gateway proxy request to a UserListRequest.
func UserListRequestFromAPIGatewayProxyRequest(
	req events.APIGatewayProxyRequest,
) (UserListRequest, error) {
	limit, err := stringToInt32Ptr(req.QueryStringParameters["limit"])
	if err != nil {
		return UserListRequest{}, fmt.Errorf("failed to parse limit query param: %w", err)
	}
	return UserListRequest{
		UserID:          req.PathParameters["user_id"],
		TimestampPrefix: req.PathParameters["ts"],
		Origin:          req.QueryStringParameters["origin"],
		OperationType:   req.QueryStringParameters["operation_type"],
		After:           req.QueryStringParameters["after"],
		Limit:           limit,
	}, nil
}

// Validate validates the request.
func (req UserListRequest) Validate() error {
	return validator.New().Struct(req)
}

// ToExpression converts the request to a DynamoDB expression.
func (req UserListRequest) ToExpression() (expression.Expression, error) {
	builder := expression.NewBuilder()
	keyCond := expression.Key("user_id").
		Equal(expression.Value(req.UserID)).
		And(expression.Key("ts").BeginsWith(req.TimestampPrefix))
	builder = builder.WithKeyCondition(keyCond)

	if req.Origin != "" {
		builder = builder.WithFilter(expression.Name("origin").Equal(expression.Value(req.Origin)))
	}

	if req.OperationType != "" {
		builder = builder.WithFilter(
			expression.Name("operation_type").Equal(expression.Value(req.OperationType)),
		)
	}
	return builder.Build()
}

// ToQueryInput converts the request to a DynamoDB query input.
// It also decodes the after cursor.
func (req UserListRequest) ToQueryInput(tableName string) (*dynamodb.QueryInput, error) {
	expr, err := req.ToExpression()
	if err != nil {
		return nil, fmt.Errorf("failed to make filter expression: %w", err)
	}

	after, err := TransactionPKFromBase64(req.After)
	if err != nil {
		return nil, fmt.Errorf("failed to decode last evaluated key: %w", err)
	}

	return &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
		Limit:                     req.Limit,
		ExclusiveStartKey:         after.ToAttributes(),
	}, nil
}
