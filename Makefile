.PHONY: build transactions-table clean rebuild test deploy

transactions-table:
	./scripts/create_test_table.sh

clean:
	rm -rf ./aws-sam

build:
	sam build

rebuild: clean build

test: transactions-table
	docker-compose start
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./...
	docker-compose stop

test-coverage:
	docker-compose start
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./... -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	docker-compose stop

deploy: rebuild
	sam deploy --guided
