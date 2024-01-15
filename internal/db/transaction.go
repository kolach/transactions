package db

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// TransactionPK represents the primary key of a transaction.
type TransactionPK struct {
	UserID    string `json:"user_id,omitempty" dynamodbav:"user_id"`
	Timestamp string `json:"ts,omitempty"      dynamodbav:"ts"`
}

// TransactionPKFromAttributes converts DynamoDB AttributeValue map to a TransactionPK.
func TransactionPKFromAttributes(attrs map[string]types.AttributeValue) TransactionPK {
	if attrs == nil {
		return TransactionPK{}
	}

	return TransactionPK{
		UserID:    attrs["user_id"].(*types.AttributeValueMemberS).Value,
		Timestamp: attrs["ts"].(*types.AttributeValueMemberS).Value,
	}
}

// TransactionPKFromBase64 converts a base64 encoded string to a TransactionPK.
func TransactionPKFromBase64(s string) (TransactionPK, error) {
	if s == "" {
		return TransactionPK{}, nil
	}

	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return TransactionPK{}, fmt.Errorf("failed to decode PK: %w", err)
	}
	var pk TransactionPK
	if err := json.Unmarshal(b, &pk); err != nil {
		return TransactionPK{}, fmt.Errorf("failed to unmarshal PK: %w", err)
	}
	return pk, nil
}

// ToAttributes converts a TransactionPK to a DynamoDB AttributeValue map.
func (pk TransactionPK) ToAttributes() map[string]types.AttributeValue {
	empty := TransactionPK{}
	if pk == empty {
		return nil
	}

	return map[string]types.AttributeValue{
		"user_id": &types.AttributeValueMemberS{Value: pk.UserID},
		"ts":      &types.AttributeValueMemberS{Value: pk.Timestamp},
	}
}

// ToBase64 converts a TransactionPK to a base64 encoded string.
func (pk TransactionPK) ToBase64() (string, error) {
	if pk == (TransactionPK{}) {
		return "", nil
	}

	b, err := json.Marshal(pk)
	if err != nil {
		return "", fmt.Errorf("failed to encode PK: %w", err)
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

// Transaction represents a transaction model
type Transaction struct {
	UserID        string  `json:"user_id"        dynamodbav:"user_id"        validate:"required"`
	Timestamp     string  `json:"ts"             dynamodbav:"ts"             validate:"required"`
	ID            string  `json:"tr_id"                                      validate:"required"       danamodbav:"tr_id"`
	Origin        string  `json:"origin"         dynamodbav:"origin"         validate:"required"`
	OperationType string  `json:"operation_type" dynamodbav:"operation_type" validate:"required"`
	Amount        float64 `json:"amount"         dynamodbav:"amount"         validate:"required,gte=0"`
}

// Timestamp returns the current timestamp in ISO 8601 format.
func Timestamp() string {
	return time.Now().Format("2006-01-02T15:04:05.999999Z")
}

// Convert tr to DynamoDB AttributeValue map
func (tr *Transaction) SetDefaults() {
	if tr.ID == "" {
		tr.ID = uuid.New().String()
	}
	if tr.Timestamp == "" {
		tr.Timestamp = Timestamp()
	}
}

// Validate validates the transaction
func (tr Transaction) Validate() error {
	// Create a new validator instance.
	v := validator.New()
	return v.Struct(&tr)
}
