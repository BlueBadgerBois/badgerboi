package db

import (
	"errors"
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

func TriggerFromUserAndStockSym(dbw *DBW, userId uint, stockSym string, trigType string ) (*Trigger, error) {
	var trig Trigger

	t := Trigger {
		UserID: userId,
		StockSym: stockSym,
		Type: trigType,
	}

	if dbw.Conn.First(&trig, &t).RecordNotFound() {
		return nil, errors.New("Error: " + trigType + " trigger not found for user " + strconv.FormatUint(uint64(userId), 10))
	}

	return &trig, nil
}
