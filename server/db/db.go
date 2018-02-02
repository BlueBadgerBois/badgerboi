package db

import (
	"log"
	"time"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DBW struct {
	Conn *gorm.DB
}

type DBTime struct {
	Now time.Time
}

func (db *DBW) Init() {
	db.connectWithRetries()
	db.autoMigrate()
	db.migrate()
}

func (db *DBW) connectWithRetries() {
	log.Println("Connecting to postgres.")
	conn, err := gorm.Open("postgres", "host=db user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	db.Conn = conn
	if err != nil {
		log.Println(err)
	}
	log.Println("Connected to postgres.")
}

func (db *DBW) Begin() *DBW {
	tx := db.Conn.Begin()
	return &DBW{ Conn: tx }
}

func (tx *DBW) Commit() {
	tx.Conn.Commit()
}

func (tx *DBW) Rollback() {
	tx.Conn.Rollback()
}

// defer this
func (db *DBW) CleanUp() {
	db.Conn.Close()
}

// automatically run any migrations that are needed (doesn't do all of them. See docs.
func (db *DBW) autoMigrate() {
	log.Println("Auto-migrating.")
	db.Conn.AutoMigrate(&User{})
	db.Conn.AutoMigrate(&Transaction{})
	db.Conn.AutoMigrate(&StockHolding{})
	db.Conn.AutoMigrate(&LogItem{})
	db.Conn.AutoMigrate(&Trigger{})
	log.Println("Finished auto-migrating.")
}

func (db *DBW) migrate() {
	// migrations go here
}

func (db *DBW) GetCurrentTime() time.Time {
	currentTime := DBTime{}
	db.Conn.Raw("select now from now();").Scan(&currentTime)
	return currentTime.Now
}
