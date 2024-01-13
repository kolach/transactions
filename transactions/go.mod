require (
	github.com/aws/aws-lambda-go v1.36.1
	github.com/aws/aws-sdk-go-v2 v1.24.1
	github.com/aws/aws-sdk-go-v2/config v1.26.3
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.12.14
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.26.8
	github.com/google/uuid v1.5.0
)

replace gopkg.in/yaml.v2 => gopkg.in/yaml.v2 v2.2.8

module transactions

go 1.16
