package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Client represents a DynamoDB client to create and fetch transactions
type Client struct {
	c     *dynamodb.Client
	table string
}

// Create creates a transaction
func (c *Client) Create(ctx context.Context, t *Transaction) error {
	t.SetDefaults()

	if err := t.Validate(); err != nil {
		return err
	}

	av, err := attributevalue.MarshalMap(t)
	if err != nil {
		return err
	}
	_, err = c.c.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      av,
	})

	return err
}

// Delete deletes a transaction
func (c *Client) Delete(ctx context.Context, t Transaction) error {
	_, err := c.c.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{
				Value: t.UserID,
			},
			"ts": &types.AttributeValueMemberS{
				Value: t.Timestamp,
			},
		},
	})
	return err
}

// DeleteAll deletes all transactions
func (c *Client) DeleteAll(ctx context.Context) error {
	transactions, err := c.Scan(ctx)
	if err != nil {
		return err
	}
	for _, t := range transactions {
		err = c.Delete(ctx, t)
		if err != nil {
			return err
		}
	}
	return nil
}

// Query lists transactions matching the request query
func (c *Client) Query(ctx context.Context, req UserListRequest) (ListResponse, error) {
	if err := req.Validate(); err != nil {
		return ListResponse{}, err
	}

	input, err := req.ToQueryInput(c.table)
	if err != nil {
		return ListResponse{}, fmt.Errorf("failed to make query input: %w", err)
	}

	res, err := c.c.Query(ctx, input)
	if err != nil {
		return ListResponse{}, err
	}

	log.Printf("last evaluated key is: %v", res.LastEvaluatedKey)

	var transactions []Transaction
	if err := attributevalue.UnmarshalListOfMaps(res.Items, &transactions); err != nil {
		return ListResponse{}, fmt.Errorf(
			"failed to decode dynamodb attributes into go struct: %w",
			err,
		)
	}

	resp := ListResponse{Items: transactions}

	pk := TransactionPKFromAttributes(res.LastEvaluatedKey)
	if resp.Cursor, err = pk.ToBase64(); err != nil {
		return ListResponse{}, err
	}

	return resp, nil
}

// Scan performs scan across all partitions
func (c *Client) Scan(ctx context.Context) ([]Transaction, error) {
	res, err := c.c.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(c.table),
	})
	if err != nil {
		return nil, err
	}
	var transactions []Transaction
	err = attributevalue.UnmarshalListOfMaps(res.Items, &transactions)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

// connect connects to DynamoDB
func connect() *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(
				func(o *retry.StandardOptions) {
					o.MaxAttempts = 5
					o.MaxBackoff = 5 * time.Second
				})
		}),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(
			func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID && os.Getenv("LOCAL_DYNAMODB_URL") != "" {
					return aws.Endpoint{
						URL:           os.Getenv("LOCAL_DYNAMODB_URL"),
						SigningRegion: region,
					}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			})),
	)
	if err != nil {
		panic(err)
	}

	return dynamodb.NewFromConfig(cfg)
}

// NewClient creates a new DynamoDB client
func NewClient() *Client {
	dynamodbClient := connect()
	tableName, _ := os.LookupEnv("TABLE_NAME")
	if tableName == "" {
		tableName = "Transactions"
	}

	return &Client{
		c:     dynamodbClient,
		table: tableName,
	}
}
