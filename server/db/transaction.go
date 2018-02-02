package db

import (
	"errors"
	"github.com/jinzhu/gorm"
)

// If a transaction explires, create a new transaction record
type Transaction struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	Type string // buy/sell
	State string // possible values: complete, pending, cancelled
	StockSymbol string // stock to buy/sell
	AmountInCents uint // amount of the stock to buy/sell. change this to cents
	QuotedStockPrice uint // Quoted price of stock in cents (valid until 60 seconds after created_at)
}

func BuildBuyTransaction(user *User) *Transaction {
	buyTransaction := Transaction{
		UserID: user.ID,
		Type: "buy",
		State: "pending",
	}

	return &buyTransaction
}

func BuildSellTransaction(user *User) *Transaction {
	sellTransaction := Transaction{
		UserID: user.ID,
		Type: "sell",
		State: "pending",
	}

	return &sellTransaction
}

func (transaction *Transaction) Cancel(db *DBW) {
	transaction.State = "cancelled"
	db.Conn.Save(transaction)
}

func NewestPendingTransactionForUser(db *DBW, user *User, txType string) (*Transaction, error) {
	transaction := Transaction{}

	notFound := db.Conn.
	Order("created_at desc").
	Where("user_id = ?", user.ID).
	Where("type = ?", txType).
	Where("state = ?", "pending").
	First(&transaction).
	RecordNotFound()

	if notFound {
		return &transaction, errors.New("No pending " + txType + " transactions found for user " + user.Username)
	}

	return &transaction, nil
}

