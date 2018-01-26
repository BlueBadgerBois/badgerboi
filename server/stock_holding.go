package main

import (
	"github.com/jinzhu/gorm"
)

type StockHolding struct {
	gorm.Model
	UserID uint `gorm:"index"`// foreign key to users table.
	StockSymbol string // Stock represented by this holding
	Number uint // Number of the stock held by the user
}
