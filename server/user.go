package main

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string
	CurrentMoney uint
	Transactions []Transaction
	StockHoldings []StockHolding
}
