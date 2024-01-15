.PHONY: build

transactions-table:
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


clean:
	rm -rf ./aws-sam

build:
	sam build

rebuild: clean build

start-dynamodb:
	docker-compose start

test: start-dynamodb transactions-table
	docker-compose start
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./...
	docker-compose stop

test-coverage:
	docker-compose start
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./... -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	docker-compose stop

deploy: build
	sam deploy --guided
