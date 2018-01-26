package main

import (
	"errors"
	"github.com/jinzhu/gorm"
)

// If a transaction explires, create a new transaction record
type Transaction struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	Type string // buy/sell
	State string // possible values: completed, pending, cancelled, expired
	StockSymbol string // stock to buy/sell
	AmountInCents uint // amount of the stock to buy/sell. change this to cents
	QuotedStockPrice uint // Quoted price of stock in cents (valid until 60 seconds after created_at)
}

func buildBuyTransaction(user *User) Transaction {
	buyTransaction := Transaction{
		UserID: user.ID,
		Type: "buy",
		State: "pending",
	}

	return buyTransaction
}

// func (db *DB) userFromUsername(username string) (*User, error) {
// 	user := User{}
// 	if db.conn.Where(&User{Username: username}).First(&user).RecordNotFound() {
// 		return &user, errors.New("User not found!")
// 	}
// 	return &user, nil
// }
func (db *DB) newestTransactionForUser(user *User) (*Transaction, error) {
	transaction := Transaction{}

	notFound := db.conn.
	Order("created_at desc").
	Where("user_id = ?", user.ID).
	Where("state = ?", "pending").
	Where("created_at > now() - interval '1 minute'"). // only select transactions that are less than a minute old
	First(&transaction).
	RecordNotFound()

	if notFound {
		return &transaction, errors.New("No pending transactions found for user " + user.Username)
	}

	return &transaction, nil
}

