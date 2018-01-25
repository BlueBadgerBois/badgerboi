package main

import (
  "github.com/jinzhu/gorm"
)


type Trigger struct {
  gorm.Model
  UserID uint `gorm:"index"`
  StockSym string
  BuyAmount uint
  PriceThreshold uint
}
