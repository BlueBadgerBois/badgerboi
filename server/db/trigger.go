package db

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

func BuildBuyTrigger(user *User) Trigger {
	buyTrigger := Trigger{
		UserID: user.ID,
		Type: "buy",
	}
  return buyTrigger
}

func BuildSellTrigger(user *User) Trigger {
  sellTrigger := Trigger{
    UserID: user.ID,
    Type: "sell",
  }
  return sellTrigger
}
