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
	db.migrate()
}

func (db *DB) connectWithRetries() {
	log.Println("Connecting to postgres.")
	conn, err := gorm.Open("postgres", "host=db user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	db.conn = conn
	if err != nil {
		log.Println(err)
	}
	log.Println("Connected to postgres.")
}

// defer this
func (db *DB) cleanUp() {
	db.conn.Close()
}

// automatically run any migrations that are needed (doesn't do all of them. See docs.
func (db *DB) autoMigrate() {
	log.Println("Auto-migrating.")
	db.conn.AutoMigrate(&User{})
	db.conn.AutoMigrate(&Transaction{})
	db.conn.AutoMigrate(&StockHolding{})
	db.conn.AutoMigrate(&LogItem{})
	db.conn.AutoMigrate(&Trigger{})
	log.Println("Finished auto-migrating.")
}

func (db *DB) migrate() {
	// migrations go here
}
