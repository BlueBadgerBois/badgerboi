package main

import (
	"log"
	// "os"
	// "time"

	"github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/postgres"
)

const MAX_RETRIES = 12
const RETRY_DELAY_SECS = 2

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

func (db *DB) autoMigrate() {
	 db.conn.AutoMigrate(&User{})
}

// func (db *DB) connectWithRetries(host string, keyspace string) {
// 	db.host = host
// 	db.cluster = gocql.NewCluster(db.host)
// 	db.cluster.Keyspace = keyspace
// 	db.cluster.Consistency = gocql.Quorum

// 	session, sess_err := db.cluster.CreateSession()

// 	retries := 0

// 	// Wait until cassandra is up
// 	for sess_err != nil {
// 		if retries > MAX_RETRIES {
// 			log.Fatal(sess_err)
// 			os.Exit(1)
// 		}

// 		time.Sleep(RETRY_DELAY_SECS * time.Second)
// 		log.Println("Error connecting to cassandra. Retrying...")
// 		session, sess_err = db.cluster.CreateSession()
// 		retries++
// 	}

// 	log.Println("Connected to cassandra successfully.")

// 	db.session = session
// }

// func (db *DB) initSchema() {
// 	log.Println("Initializing schema")

// 	// Add all schemas here
// 	db.createTable(userSchema)
// }

// func (db *DB) createTable(table string) {
// 	if err := db.session.Query(table).RetryPolicy(nil).Exec(); err != nil {
// 		log.Printf("error creating table table=%q err=%v\n", table, err)
// 	}
// }

