package test

import (
	"github.com/Pallinder/go-randomdata"
	"github.com/google/uuid"
	"github.com/kolach/go-factory"

	"transactions/internal/db"
)

func amount() float64 {
	return float64(randomdata.Number(1000))
}

var TransactionFactory = factory.NewFactory(
	db.Transaction{},
	factory.Use("nick", "john", "james", "foo", "bar").For("UserID"),
	factory.Use(uuid.NewString).For("ID"),
	factory.Use("credit", "debit").For("OperationType"),
	factory.Use("web", "mobile", "ios", "android", "desktop").For("Origin"),
	factory.Use(db.Timestamp).For("Timestamp"),
	factory.Use(amount).For("Amount"),
)
