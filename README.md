# Transactions project

## System design

The Transactions project is an AWS SAM project that delegates all transaction APIs to a serverless Lambda function that uses DynamoDB service to store and retrieve transaction data.

```bash
+------------------+      +---------------------+      +-------------------+
|                  |      |                     |      |                   |
|   API Gateway    +----->+   AWS Lambda Func   +----->+    DynamoDB       |
|                  |      |                     |      |                   |
+------------------+      +---------------------+      +-------------------+
```

The main benefit of this approach is that it scales perfectly horizontally at both the application and database levels, and it is charged only when it is being used.

## Data model

The system design is centered around a multi-tenant application where a multitude of users are entering their transactions and have a way to monitor them.

Below is the transaction data model:

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

The User ID (user_id) is set as the partition key in an attempt to evenly distribute incoming transactions across DynamoDB shards. The Timestamp (ts) is used as a sort key.

Together, they form a primary key. Should there be any concern that transactions with the same user ID and timestamp (in microseconds) might exist, the design can accommodate this by making the sort key a composite key of the timestamp and the unique transaction ID (tr_id).
User ID (user_id) is set as partition key in attempt to evenly distribute incoming transactions across dynamodb shards.
Timstamp (ts) is used as a sort key.

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
│   └── populate                <-- CLI tool to send POST random transaction requests to AWS transactions API endpoint
│       └── main.go             <-- CLI tool code
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


The prerequisites for deploying the project are quite common. If you've already had experience with AWS services, you most likely have everything installed.

For detailed instructions, follow the link: [link](https://docs.aws.amazon.com/serverless-application-model/latest/developerguide/prerequisites.html) for details
and it will guide you through the following steps::

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

Answer the questions to complete the deployment. Pay close attention to the final output parameters to obtain your TransactionsAPI URL.


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

Let's use the API URL obtained from the previous chapter to create and list transactions. We'll start with creation. To create a new transaction, send a POST request:

```bash
curl -X POST -H "Content-Type: application/json" -d '{"user_id":"john", "amount":1,"origin":"desktop", "operation_type":"credit"}' $TRANSACTIONS_API
```
```bash
curl -X POST -H "Content-Type: application/json" -d '{"user_id":"john", "amount":1,"origin":"desktop", "operation_type":"debit"}' $TRANSACTIONS_API
```

You will need to define the TRANSACTIONS_API endpoint as an environment variable, or alternatively, use the actual URL directly. Utilize the example provided above to create a few more transactions.

Now, let's move on to listing transactions. To list transactions, we need to make a GET request and use a partition key (user_id) and a sort key (ts) as path parameters:

```shell
$TRANSACTIONS_API/{user_id}/{ts}
```

The timestamp is used as a prefix (using the 'begins_with' filter). Therefore, to retrieve all transactions belonging to john dated in the year 2024, we need to send the following request:


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
The returned JSON object contains two fields:

items - an array of transactions.
cursor - a string that, if returned, represents a cursor for the next page. See below for details.
The list request supports the following optional query parameters (as URL query parameters):

origin: string
operation_type: string
limit: number (to limit the maximum number of returned records)
after: string (a cursor pagination parameter to supply in order to get the next page)
Use the cursor attribute from the returned object to access the next page of data.

Now, let's make a sample query to retrieve credit transactions for `john`:

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

It's expected for AWS to scale Lambdas according to the configured concurrency parameter, but this depends on the settings of the AWS account. For example, my account currently has a concurrency limit of only 10. This limitation restricts the scaling of Lambda instances to no more than 10, and impact performance as requests may be throttled when all Lambdas are active and busy.

Using the timestamp prefix as a parameter in the URL path proved to be suboptimal for filtering transactions by minutes. This is because adding the : symbol, which is necessary for minute-level filtering, can cause issues in the URL path. However, it works fine in query parameters if replaced with %3A.

The handlers are equipped with basic tests. Although they call the DynamoDB internal package, they would benefit from having their own dedicated test package.

Need to find a way to better integrate Swagger into the project.

## Running tests

`docker` and `docker-compose` is required to run tests. As they are used to run local version of dynamodb.
Use following command to run tests:

```shell
make test
```
