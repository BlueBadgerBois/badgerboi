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

func (dbw *DBW) Init() {
	dbw.connectWithRetries()

	dbw.Conn.DB().SetMaxOpenConns(400)
	dbw.Conn.DB().SetMaxIdleConns(400)

	dbw.autoMigrate()
	dbw.migrate()
}

func (dbw *DBW) connectWithRetries() {
	log.Println("Connecting to postgres.")
	// conn, err := gorm.Open("postgres", "host=db_proxy port=6432 user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	conn, err := gorm.Open("postgres", "host=db port=5432 user=badgerboi dbname=badgerboi sslmode=disable password=badgerboi")
	dbw.Conn = conn
	if err != nil {
		log.Println(err)
	}
	log.Println("Connected to postgres.")
}

func (dbw *DBW) Begin() *DBW {
	tx := dbw.Conn.Begin()
	return &DBW{ Conn: tx }
}

func (tx *DBW) Commit() {
	tx.Conn.Commit()
}

func (tx *DBW) Rollback() {
	tx.Conn.Rollback()
}

// defer this
func (dbw *DBW) CleanUp() {
	dbw.Conn.Close()
}

// automatically run any migrations that are needed (doesn't do all of them. See docs.
func (dbw *DBW) autoMigrate() {
	log.Println("Auto-migrating.")
	dbw.Conn.AutoMigrate(&User{})
	dbw.Conn.AutoMigrate(&Transaction{})
	dbw.Conn.AutoMigrate(&StockHolding{})
	dbw.Conn.AutoMigrate(&LogItem{})
	dbw.Conn.AutoMigrate(&Trigger{})
	dbw.Conn.AutoMigrate(&TransactionNum{})
	log.Println("Finished auto-migrating.")
}

func (dbw *DBW) migrate() {
	// migrations go here
}

func (dbw *DBW) GetCurrentTime() time.Time {
	currentTime := DBTime{}
	dbw.Conn.Raw("select now from now();").Scan(&currentTime)
	return currentTime.Now
}
