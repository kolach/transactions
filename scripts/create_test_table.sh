#!/bin/bash

# Disable cli pager
# export AWS_PAGER=""

TABLE_NAME="Transactions"
ENDPOINT_URL="http://localhost:8000" # URL of your local DynamoDB instance

# Timeout and interval in seconds
TIMEOUT=30
INTERVAL=5
ELAPSED=0

echo "Waiting for DynamoDB to start on port 8000..."

# Function to check if DynamoDB is running
is_dynamodb_running() {
    # Using lsof to check if the port is in use
    if lsof -i :8000 | grep -q LISTEN; then
        return 0 # DynamoDB is running
    fi
    return 1 # DynamoDB is not running
}

# Loop until the service is up or timeout is reached
while ! is_dynamodb_running; do
    if [ "$ELAPSED" -ge "$TIMEOUT" ]; then
        echo "Timeout reached. DynamoDB did not start."
        exit 1
    fi

    echo "Waiting for DynamoDB to start..."
    sleep $INTERVAL
    ELAPSED=$((ELAPSED+INTERVAL))
done

echo "DynamoDB is up and running on port 8000."

# Check if the table exists
if aws dynamodb describe-table --no-cli-pager --table-name $TABLE_NAME --endpoint-url $ENDPOINT_URL > /dev/null 2>&1; then
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
