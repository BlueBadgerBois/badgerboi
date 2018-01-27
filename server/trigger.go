package main

import (
  "github.com/jinzhu/gorm"
)

type Trigger struct {
  gorm.Model
  UserID uint `gorm:"index"`
  Type string
  StockSym string
  Amount uint // cents
  NumStocks uint // for sell triggers only, leave empty for buy
  PriceThreshold uint
}

func buildBuyTrigger(user *User) Trigger {
	buyTrigger := Trigger{
		UserID: user.ID,
		Type: "buy",
	}
  return buyTrigger
}

func buildSellTrigger(user *User) Trigger {
  sellTrigger := Trigger{
    UserID: user.ID,
    Type: "sell",
  }
  return sellTrigger
}
