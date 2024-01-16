.PHONY: build transactions-table clean rebuild test deploy

clean:
	rm -rf ./aws-sam

build:
	sam build

rebuild: clean build

test:
	docker-compose pull
	docker-compose up -d
	./scripts/create_test_table.sh
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./...
	docker-compose down

test-coverage:
	docker-compose pull
	docker-compose up -d
	./scripts/create_test_table.sh
	LOCAL_DYNAMODB_URL=http://localhost:8000 go test ./... -coverprofile cover.out
	go tool cover -html=cover.out -o cover.html
	docker-compose down

deploy: rebuild
	sam deploy --guided
