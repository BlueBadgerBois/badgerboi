package main

import (
	"log"
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DBW struct {
	conn *gorm.DB
}

type DBTime struct {
	Now time.Time
}

func (db *DBW) init() {
	db.connectWithRetries()
	db.autoMigrate()
	db.migrate()
}

func (db *DBW) connectWithRetries() {
	log.Println("Connecting to postgres.")
	conn, err := gorm.Open("postgres", "host=db user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	db.conn = conn
	if err != nil {
		log.Println(err)
	}
	log.Println("Connected to postgres.")
}

func (db *DBW) Begin() *DBW {
	tx := db.conn.Begin()
	return &DBW{ conn: tx }
}

func (tx *DBW) Commit() {
	tx.conn.Commit()
}

func (tx *DBW) Rollback() {
	tx.conn.Rollback()
}

// defer this
func (db *DBW) cleanUp() {
	db.conn.Close()
}

// automatically run any migrations that are needed (doesn't do all of them. See docs.
func (db *DBW) autoMigrate() {
	log.Println("Auto-migrating.")
	db.conn.AutoMigrate(&User{})
	db.conn.AutoMigrate(&Transaction{})
	db.conn.AutoMigrate(&StockHolding{})
	db.conn.AutoMigrate(&LogItem{})
	db.conn.AutoMigrate(&Trigger{})
	log.Println("Finished auto-migrating.")
}

func (db *DBW) migrate() {
	// migrations go here
}

func (db *DBW) getCurrentTime() time.Time {
	currentTime := DBTime{}
	db.conn.Raw("select now from now();").Scan(&currentTime)
	return currentTime.Now
}
