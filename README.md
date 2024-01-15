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

### Local development

**Invoking function locally through local API Gateway**

```bash
sam local start-api
```

If the previous command ran successfully you should now be able to hit the following local endpoint to invoke your function `http://localhost:3000/hello`

**SAM CLI** is used to emulate both Lambda and API Gateway locally and uses our `template.yaml` to understand how to bootstrap this environment (runtime, where the source code is, etc.) - The following excerpt is what the CLI will read in order to initialize an API and its routes:

```yaml
...
Events:
    HelloWorld:
        Type: Api # More info about API Event Source: https://github.com/awslabs/serverless-application-model/blob/master/versions/2016-10-31.md#api
        Properties:
            Path: /hello
            Method: get
```

## Packaging and deployment

AWS Lambda Golang runtime requires a flat folder with the executable generated on build step. SAM will use `CodeUri` property to know where to look up for the application:

```yaml
...
    FirstFunction:
        Type: AWS::Serverless::Function
        Properties:
            CodeUri: hello_world/
            ...
```

To deploy your application for the first time, run the following in your shell:

```bash
sam deploy --guided
```

The command will package and deploy your application to AWS, with a series of prompts:

* **Stack Name**: The name of the stack to deploy to CloudFormation. This should be unique to your account and region, and a good starting point would be something matching your project name.
* **AWS Region**: The AWS region you want to deploy your app to.
* **Confirm changes before deploy**: If set to yes, any change sets will be shown to you before execution for manual review. If set to no, the AWS SAM CLI will automatically deploy application changes.
* **Allow SAM CLI IAM role creation**: Many AWS SAM templates, including this example, create AWS IAM roles required for the AWS Lambda function(s) included to access AWS services. By default, these are scoped down to minimum required permissions. To deploy an AWS CloudFormation stack which creates or modifies IAM roles, the `CAPABILITY_IAM` value for `capabilities` must be provided. If permission isn't provided through this prompt, to deploy this example you must explicitly pass `--capabilities CAPABILITY_IAM` to the `sam deploy` command.
* **Save arguments to samconfig.toml**: If set to yes, your choices will be saved to a configuration file inside the project, so that in the future you can just re-run `sam deploy` without parameters to deploy changes to your application.

You can find your API Gateway Endpoint URL in the output values displayed after deployment.

### Testing

We use `testing` package that is built-in in Golang and you can simply run the following command to run our tests:

```shell
cd ./hello-world/
go test -v .
```
# Appendix

### Golang installation

Please ensure Go 1.x (where 'x' is the latest version) is installed as per the instructions on the official golang website: https://golang.org/doc/install

A quickstart way would be to use Homebrew, chocolatey or your linux package manager.

#### Homebrew (Mac)

Issue the following command from the terminal:

```shell
brew install golang
```

If it's already installed, run the following command to ensure it's the latest version:

```shell
brew update
brew upgrade golang
```

#### Chocolatey (Windows)

Issue the following command from the powershell:

```shell
choco install golang
```

If it's already installed, run the following command to ensure it's the latest version:

```shell
choco upgrade golang
```

## Bringing to the next level

Here are a few ideas that you can use to get more acquainted as to how this overall process works:

* Create an additional API resource (e.g. /hello/{proxy+}) and return the name requested through this new path
* Update unit test to capture that
* Package & Deploy

Next, you can use the following resources to know more about beyond hello world samples and how others structure their Serverless applications:

* [AWS Serverless Application Repository](https://aws.amazon.com/serverless/serverlessrepo/)
