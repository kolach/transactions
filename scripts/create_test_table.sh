#!/bin/bash

TABLE_NAME="Transactions"
ENDPOINT_URL="http://localhost:8000" # URL of your local DynamoDB instance

# Check if the table exists
if aws dynamodb describe-table --table-name $TABLE_NAME --endpoint-url $ENDPOINT_URL 2>/dev/null; then
  echo "Table $TABLE_NAME already exists."
else
  # Create the table
	aws dynamodb create-table \
		--table-name Transactions \
		--attribute-definitions \
			AttributeName=user_id,AttributeType=S \
			AttributeName=ts,AttributeType=S \
		--key-schema \
			AttributeName=user_id,KeyType=HASH \
			AttributeName=ts,KeyType=RANGE \
		--billing-mode PAY_PER_REQUEST \
		--endpoint-url http://localhost:8000 \
		--stream-specification StreamEnabled=true,StreamViewType=NEW_IMAGE

  echo "Table $TABLE_NAME created."
fi
