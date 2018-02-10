package db

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type TransactionNum struct {
	gorm.Model
	TransactionId int `gorm:"AUTO_INCREMENT;unique;not null"`
}

func NewTxNum(dbw *DBW) uint {
	txNum := TransactionNum{}
	dbw.Conn.Create(&txNum)
	return txNum.ID
}

func FirstTxNum(dbw *DBW) (uint, error){
	txNum := TransactionNum{}
	if dbw.Conn.First(&txNum).RecordNotFound() {
		return txNum.ID, errors.New("no txnum found")
	}
	return txNum.ID, nil
}
