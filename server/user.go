package main

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string
	Current_money uint
	Transactions []Transaction
	StockHoldings []StockHolding
}
