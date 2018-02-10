package db

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

func BuildStockHolding(dbw *DBW, user *User, stockSymbol string) (*StockHolding, error) {
	holdingQuery := StockHolding{UserID: user.ID, StockSymbol: stockSymbol}
	holding := StockHolding{}
	notFound := dbw.Conn.
	Where(holdingQuery).
	First(&holding).
	RecordNotFound()

	if notFound {
		errMsg := "No stock holdings for symbol "+ stockSymbol + " found for user " + user.Username
		return &holding, errors.New(errMsg)
	}

	return &holding, nil
}

func StockHoldingFromUserAndStockSym(dbw *DBW, userId uint, stockSymbol string) (*StockHolding, error) {
	var stockHolding StockHolding
	
	s := StockHolding{
		UserID: userId,
		StockSymbol: stockSymbol,
	}
	
	if dbw.Conn.First(&stockHolding, &s).RecordNotFound() {
		return nil, errors.New("Error: " + stockSymbol + " holding not found for user " + string(userId))
	}

	return &stockHolding, nil
}

func (holding *StockHolding) Sufficient(targetNumber uint) bool {
	return holding.Number >= targetNumber
}

// assumes you won't withdraw over the limit
func (holding *StockHolding) Withdraw(dbw *DBW, numToWithdraw uint) {
	holding.Number -= numToWithdraw
	dbw.Conn.Save(holding)
}
