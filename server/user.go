package main

import (
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string
	CurrentMoney uint
	Transactions []Transaction
	StockHoldings []StockHolding
}

func userFromUsername(username string) *User {
	u := User{Username: username}

	var user User
	db.conn.FirstOrCreate(&user, &u)
	return &user
}
