package main

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type StockHolding struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	StockSymbol string // Stock represented by this holding
	Number uint // Number of the stock held by the user
}

func buildStockHolding(db *DBW, user *User, stockSymbol string) (*StockHolding, error) {
	holdingQuery := StockHolding{UserID: user.ID, StockSymbol: stockSymbol}
	holding := StockHolding{}
	notFound := db.conn.
	Where(holdingQuery).
	First(&holding).
	RecordNotFound()

	if notFound {
		errMsg := "No stock holdings for symbol "+ stockSymbol + " found for user " + user.Username
		return &holding, errors.New(errMsg)
	}

	return &holding, nil
}

func (holding *StockHolding) Sufficient(targetNumber uint) bool {
	return holding.Number >= targetNumber
}

// assumes you won't withdraw over the limit
func (holding *StockHolding) Withdraw(db *gorm.DB, numToWithdraw uint) {
	holding.Number -= numToWithdraw
	db.Save(holding)
}
