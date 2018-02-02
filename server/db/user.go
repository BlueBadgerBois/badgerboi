package db

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
	db.Conn.FirstOrCreate(&user, &u)
	return &user
}

func UserFromUsername(db *DBW, username string) (*User, error) {
	user := User{}
	if db.Conn.Where(&User{Username: username}).First(&user).RecordNotFound() {
		return &user, errors.New("User " + username + " not found!")
	}
	return &user, nil
}

func (user *User) HasEnoughMoney(targetAmount uint) bool {
	return user.CurrentMoney >= targetAmount
}

func (user *User) DepositMoney(db *DBW, moneyInCents uint) {
	user.CurrentMoney += moneyInCents
	db.Conn.Save(user)
}