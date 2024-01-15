# Transactions project

## System design

Transactions project is AWS SAM project that deligates all transactions API to serverless
lambda funtion interaction with DynamoDB table.

```bash
+------------------+      +---------------------+      +-------------------+
|                  |      |                     |      |                   |
|   API Gateway    +----->+   AWS Lambda Func   +----->+    DynamoDB       |
|                  |      |                     |      |                   |
+------------------+      +---------------------+      +-------------------+
```

The main benefit of this approach is that it perfectly scheduled horizontally on application
and database layers and is charged only when it is used.

## Data model

The system design is centerd around multi tenant application where multitude of users are entering their
transations and have a way to monitor them.

Bellow is the transaction data model:

```bash
+------------------------------------------------------------------------------------------------+
|                                     Transactions                                               |
+-----------------+-------------------------------+----------+----------+----------------+-------+
| user_id         | ts                            | tr_id    | origin   | operation_type | amount|
| (Partition Key) | (Sort Key in ISO 8601)        |          |          |                |       |
+-----------------+-------------------------------+----------+----------+----------------+-------+
| "User123"       | "2024-01-15T10:00:00.822373Z" | "Trx456" | "Desktop"| "Debit"        | 100.0 |
| "User123"       | "2024-01-16T11:30:00.183374Z" | "Trx457" | "iOS"    | "Credit"       | 50.0  |
| "User456"       | "2024-01-15T15:45:00.382415Z" | "Trx458" | "Android"| "Debit"        | 200.0 |
| ...             | ...                           | ...      | ...      | ...            | ...   |
+-----------------+-------------------------------+----------+----------+----------------+-------+

```

User ID (user_id) is set as partition in attempt to evenly distribute incoming transactions across dynamodb shards.
Timstamp (ts) is used as a sort key.

Together they form a primary key. Should there be any concern that there can exist transactions with same
user ID and timestamp (in microseconds) the design can accomodate it with making sort key as composition key
of typestamp and the unique transaction ID (tr_id).

## Project structure

These are main files for transactions with brief explanation:

```bash
.
├── Makefile                    <-- Automated tasks
├── README.md                   <-- This instructions file
├── cmd                         <-- Root directory for lambda function and cli apps
│   ├── transactions            <-- Lambda function code
│   │   ├── main.go             <-- Lambda function code
│   │   └── main-test.go        <-- Lambda function tests
│   └── populate                <-- Lambda function code
│       └── main.go             <-- Script to massively call your lambda functions to add new random transactions
├── internal                    <-- Root directory for internal packages
│   └── db                      <-- Package to work with DynamoDB (add, remove, list, scan records)
│       ├── client.go           <-- Client to perform all CRUD operations
│       ├── transaction.go      <-- Transaction data model
│       ├── query.go            <-- Query interface and convertion helpers
│       └── util.go             <-- helper functions
├── swagger.yaml                <-- API documentation
└── template.yaml               <-- SAM template file
```

## Prerequisites

The prerequisites to deploy the project are pretty common. If you are already got your hands dirty with AWS services,
most likely you already have it all installed.

Follow the [link](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/prerequisites.html) for details
and it'll guide you through the following steps:

1. Sign up for an AWS account
2. Create an IAM user account
3. Create an access key ID and secret access key
4. Install the AWS CLI
5. Use the AWS CLI to configure AWS credentials

Then install SAM CLI:
Follow this [link](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/install-sam-cli.html) for details

### Building and deploying the target

Building the target:

```shell
make build
```

Deploying the target:

```shell
make deploy
```

Answer the questions to make a deployment. Pay attention to final output parameters to grab your TransactionsAPI url.


```shell
....
Key                 TransactionsAPI
Description         API Gateway endpoint URL for Prod environment for Transactions Function
Value               https://xyzxy18zxy.execute-api.us-east-1.amazonaws.com/Prod/transactions/

```

In the example above the following URL is the transactions API URL to interact with:
```shell
https://xyzxy18zxy.execute-api.us-east-1.amazonaws.com/Prod/transactions/
```

We will need that URL in the next chapter.

## Interacting with Transaction API

Let's use the API URL that we found in previous chapter to create and list transactions. We start with creating.
To create a new transaction send a POST request:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"user_id":"john", "amount":1,"origin":"desktop", "operation_type":"credit"}' $TRANSACTIONS_API
```
```bash
curl -X POST -H "Content-Type: application/json" -d '{"user_id":"john", "amount":1,"origin":"desktop", "operation_type":"debit"}' $TRANSACTIONS_API
```

You would need to define TRANSACTIONS_API endpoint as an environment variable or use real URL instead.
Use the example above to create a few more transactions.

Now let's list transactions. To list transactions we need to make GET request and use a partition key (user_id) and sort key (ts) as a path parameters:

```shell
$TRANSACTIONS_API/{user_id}/{ts}
```

Timestamp is used as a prefix (begins_with filter). So to get all transactions that belong to `john` dated by 2024 year we need to send request:


```shell
curl -s $TRANSACTIONS_API/john/2024 | jq
{
  "items": [
    {
      "user_id": "john",
      "ts": "2024-01-15T18:18:36.819581Z",
      "tr_id": "1031391a-8188-4cd7-b9f2-5cd6c89a1974",
      "origin": "desktop",
      "operation_type": "credit",
      "amount": 1
    },
    {
      "user_id": "john",
      "ts": "2024-01-15T18:18:56.639872Z",
      "tr_id": "f8c49906-551b-4474-80e9-e953ce3edc57",
      "origin": "desktop",
      "operation_type": "debit",
      "amount": 1
    }
  ]
}
```

The returned JSON object has 2 fields:

* items - array  of transactions
* cursor - string, if returned, represents a cursor for the next page. See bellow.

The list request supports following optional query parameters (as url query parameters):

* origin: string
* operation_type: string
* limit: number (to limit maximum number of returned records)
* after: string (cursor pagination parameter to supply to get next page)

Use `cursor` attribute of returned object to get the next page of data.

Let's make a sample query to get john's credit transactions:

```bash
curl -s "$TRANSACTIONS_API/john/2024?operation_type=credit" | jq
{
  "items": [
    {
      "user_id": "john",
      "ts": "2024-01-15T18:18:36.819581Z",
      "tr_id": "1031391a-8188-4cd7-b9f2-5cd6c89a1974",
      "origin": "desktop",
      "operation_type": "credit",
      "amount": 1
    }
  ]
}
```

## Limitations and things to improve

It's supposed AWS to scale lambdas according to configured concurrency parameter. But it depends on AWS account settings.
For example my account currently only have it 10. Which does not allow me to have lambda scaled more than 10 instances.
It of cause impacts the performance as requests are throttled when all lambdas are up and busy.

Having tamestamp prefix parameter in URL path proved to be not the best option as to filter transactions by minutes
and further reuires to add `:` symbol and it's a problem to have it in URL path. It works fine in query parameters if replaced
by `%3A`.

Handlers have basic tests. But though even they are calling DynamoDB internal package, they deserve their own test package.

### Running tests

`docker` and `docker-compose` is required to run tests. As they are used to run local version of dynamodb.
Use following command to run tests:

```shell
make test
```
