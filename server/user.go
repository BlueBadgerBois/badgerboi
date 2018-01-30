package main

import (
	"errors"
	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username string
	CurrentMoney uint
	Transactions []Transaction
	StockHoldings []StockHolding
}

func  UserFromUsernameOrCreate(db *DBW, username string) *User {
	u := User{Username: username}

	var user User
	db.conn.FirstOrCreate(&user, &u)
	return &user
}

func UserFromUsername(db *DBW, username string) (*User, error) {
	user := User{}
	if db.conn.Where(&User{Username: username}).First(&user).RecordNotFound() {
		return &user, errors.New("User " + username + " not found!")
	}
	return &user, nil
}

func (user *User) HasEnoughMoney(targetAmount uint) bool {
	return user.CurrentMoney >= targetAmount
}

func (user *User) DepositMoney(db *gorm.DB, moneyInCents uint) {
	user.CurrentMoney += moneyInCents
	db.Save(user)
}
