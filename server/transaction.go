package main

import (
	"github.com/jinzhu/gorm"
)

// If a transaction explires, create a new transaction record
type Transaction struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	Type string // buy/sell
	State string // possible values: completed, pending, cancelled, expired
	StockSymbol string // stock to buy/sell
	DollarAmount uint // amount of the stock to buy/sell
	QuotedStockPrice uint // Quoted price of stock (valid until 60 seconds after created_at)
}
