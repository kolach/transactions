package test

import (
	"testing"
	"transactions/internal/db"
)

func TestFactory(t *testing.T) {
	tr := TransactionFactory.MustCreate().(*db.Transaction)
	if err := tr.Validate(); err != nil {
		t.Errorf("expected no error but got: %v", err)
	}
}
