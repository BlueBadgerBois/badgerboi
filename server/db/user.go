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

func  UserFromUsernameOrCreate(dbw *DBW, username string) *User {
	u := User{Username: username}

	var user User
	dbw.Conn.FirstOrCreate(&user, &u)
	fmt.Println("user id from newly created username: ", user.Id)
	return &user
}

func UserFromUsername(dbw *DBW, username string) (*User, error) {
	user := User{}
	if dbw.Conn.Where(&User{Username: username}).First(&user).RecordNotFound() {
		return &user, errors.New("User " + username + " not found!")
	}
	fmt.Println("user id from username: ", user.Id)
	return &user, nil
}

func (user *User) HasEnoughMoney(targetAmount uint) bool {
	return user.CurrentMoney >= targetAmount
}

func (user *User) DepositMoney(dbw *DBW, moneyInCents uint) {
	user.CurrentMoney += moneyInCents
	dbw.Conn.Save(user)
}
