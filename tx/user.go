package main

import (
	"github.com/jinzhu/gorm"
)

// const USER_TABLE_NAME = "badgerboi.user"

// // username is the PK and therefore must be unique
// var userSchema = `
// CREATE TABLE IF NOT EXISTS ` + USER_TABLE_NAME + ` (
// 	id UUID,
// 	username text,
// 	PRIMARY KEY(username)
// )`

type User struct {
	gorm.Model
	Username string
}

// func (db *DB) createUser(username string) {
// 	var u = &User{
// 		Id: gocql.TimeUUID(),
// 		Username: username,
// 	}

// 	stmt, names := qb.Insert(USER_TABLE_NAME).Columns("id", "username").ToCql()

// 	q := gocqlx.Query(db.session.Query(stmt), names).BindStruct(u)

// 	if err := q.ExecRelease(); err != nil {
// 		log.Fatal(err)
// 	}
// }

