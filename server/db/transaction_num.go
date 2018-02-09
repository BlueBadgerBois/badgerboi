package db

import (
	"github.com/jinzhu/gorm"
)

type TransactionNum struct {
	gorm.Model
	TransactionId int `gorm:"AUTO_INCREMENT;unique;not null"`
}

func NewTxNum(dbw *DBW) *TransactionNum {
	txNum := TransactionNum{}
	dbw.Conn.Create(&txNum)
	return &txNum
}