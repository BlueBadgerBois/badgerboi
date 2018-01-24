package main

import (
	"github.com/jinzhu/gorm"
)

// If a transaction explires, create a new transaction record
type StockHolding struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	StockSymbol string // stock to buy/sell
	Number uint // Number of the stock held by the user
}
