package main

import (
  "github.com/jinzhu/gorm"
)

type Trigger struct {
  gorm.Model
  UserID uint `gorm:"index"`
  Type string
  StockSym string
  Amount uint
  PriceThreshold uint
}

func buildBuyTrigger(user *User) Trigger {
	buyTrigger := Trigger{
		UserID: user.ID,
		Type: "buy",
	}
  return buyTrigger
}
