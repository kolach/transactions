package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"transactions/internal/db"
	"transactions/pkg/test"
)

func cleanUp() error {
	client := db.NewClient()
	return client.DeleteAll(context.Background())
}

func TestMain(m *testing.M) {
	if err := cleanUp(); err != nil {
		panic(err)
	}
	code := m.Run()
	if err := cleanUp(); err != nil {
		panic(err)
	}
	os.Exit(code)
}

func MustMarshalJSON(t *testing.T, i interface{}) string {
	b, err := json.Marshal(i)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestHandler(t *testing.T) {
	tr := test.TransactionFactory.MustCreate().(*db.Transaction)

	// b, _ := json.Marshal(db.ListResponse{Items: []db.Transaction{*tr}})

	testCases := []struct {
		name           string
		request        events.APIGatewayProxyRequest
		expectedBody   string
		expectedError  error
		expectedStatus int
	}{
		{
			name: "create invalid transaction without amount",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				RequestContext: events.APIGatewayProxyRequestContext{
					Identity: events.APIGatewayRequestIdentity{
						SourceIP: "",
					},
				},
				Body: `{"user_id":"john","id":"a","operation_type":"credit","origin":"web"}`,
			},
			expectedBody:   "failed to create record: Key: 'Transaction.Amount' Error:Field validation for 'Amount' failed on the 'required' tag",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  nil,
		},
		{
			name: "create transaction",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				RequestContext: events.APIGatewayProxyRequestContext{
					Identity: events.APIGatewayRequestIdentity{
						SourceIP: "",
					},
				},
				Body: MustMarshalJSON(t, tr),
			},
			expectedBody:   MustMarshalJSON(t, tr),
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
		{
			name: "list transactions not specifying partition and sort keys",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				RequestContext: events.APIGatewayProxyRequestContext{
					Identity: events.APIGatewayRequestIdentity{
						SourceIP: "",
					},
				},
			},
			expectedBody:   "failed to scan records in dynamodb: Key: 'UserListRequest.UserID' Error:Field validation for 'UserID' failed on the 'required' tag\nKey: 'UserListRequest.TimestampPrefix' Error:Field validation for 'TimestampPrefix' failed on the 'required' tag",
			expectedStatus: http.StatusInternalServerError,
			expectedError:  nil,
		},
		{
			name: "list transactions",
			request: events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				PathParameters: map[string]string{
					"user_id": tr.UserID,
					"ts":      tr.Timestamp,
				},
				RequestContext: events.APIGatewayProxyRequestContext{
					Identity: events.APIGatewayRequestIdentity{
						SourceIP: "",
					},
				},
			},
			expectedBody:   MustMarshalJSON(t, db.ListResponse{Items: []db.Transaction{*tr}}),
			expectedStatus: http.StatusOK,
			expectedError:  nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response, err := handler(testCase.request)
			if err != testCase.expectedError {
				t.Errorf("Expected error %v, but got %v", testCase.expectedError, err)
			}

			if response.Body != testCase.expectedBody {
				t.Errorf("Expected response %v, but got %v", testCase.expectedBody, response.Body)
			}

			if response.StatusCode != testCase.expectedStatus {
				t.Errorf(
					"Expected status code %v, but got %v",
					testCase.expectedStatus,
					response.StatusCode,
				)
			}
		})
	}
}
