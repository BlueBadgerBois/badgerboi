package main

import (
	"log"

	"github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

type DB struct {
	conn *gorm.DB
}

func (db *DB) init() {
	db.connectWithRetries()
	db.autoMigrate()
}

func (db *DB) connectWithRetries() {
	conn, err := gorm.Open("postgres", "host=db user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	db.conn = conn
	if err != nil {
		log.Println(err)
	}

	db.autoMigrate()
}

// defer this
func (db *DB) cleanUp() {
	db.conn.Close()
}

// automatically run any migrations that are needed (doesn't do all of them. See docs.
func (db *DB) autoMigrate() {
	 db.conn.AutoMigrate(&User{})
}
