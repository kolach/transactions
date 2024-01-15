package db

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
)

func TestUserListRequest_ToExpression(t *testing.T) {
	req := UserListRequest{
		UserID:          "123",
		TimestampPrefix: "2021",
		Origin:          "web",
		OperationType:   "create",
	}

	expectedKeyCond := expression.Key("user_id").
		Equal(expression.Value(req.UserID)).
		And(expression.Key("ts").BeginsWith(req.TimestampPrefix))

	builder := expression.NewBuilder()
	builder = builder.WithKeyCondition(expectedKeyCond)
	builder = builder.WithFilter(expression.Name("origin").Equal(expression.Value(req.Origin)))
	builder = builder.WithFilter(
		expression.Name("operation_type").Equal(expression.Value(req.OperationType)))

	expectedExpr, err := builder.Build()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	expr, err := req.ToExpression()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	keyCond := expr.KeyCondition()
	filter := expr.Filter()

	if *keyCond != *expectedExpr.KeyCondition() {
		t.Errorf("expected key condition: %s, got: %s", *expectedExpr.KeyCondition(), *keyCond)
	}

	if *filter != *expectedExpr.Filter() {
		t.Errorf("expected filter: %s, got: %s", *expectedExpr.Filter(), *filter)
	}
}
