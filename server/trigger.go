package main

import (
  "github.com/jinzhu/gorm"
)

type Trigger struct {
  gorm.Model
  UserID uint `gorm:"index"`
  Type string
  StockSym string
  Amount uint // Buy trigger: cents. Sell trigger: number of stocks
  PriceThreshold uint
}

func buildBuyTrigger(user *User) Trigger {
	buyTrigger := Trigger{
		UserID: user.ID,
		Type: "buy",
	}
  return buyTrigger
}
