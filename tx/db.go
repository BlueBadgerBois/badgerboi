package main

import (
	"log"
	"os"
	"github.com/gocql/gocql"
	"time"
)

const MAX_RETRIES = 12
const RETRY_DELAY_SECS = 2

type DB struct {
	host string
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func (db *DB) connectWithRetries(host string, keyspace string) {
	db.host = host
	db.cluster = gocql.NewCluster(db.host)
	db.cluster.Keyspace = keyspace
	db.cluster.Consistency = gocql.Quorum

	session, sess_err := db.cluster.CreateSession()

	retries := 0

	// Wait until cassandra is up
	for sess_err != nil {
		if retries > MAX_RETRIES {
			log.Fatal(sess_err)
			os.Exit(1)
		}

		time.Sleep(RETRY_DELAY_SECS * time.Second)
		log.Println("Error connecting to cassandra. Retrying...")
		session, sess_err = db.cluster.CreateSession()
		retries++
	}

	log.Println("Connected to cassandra successfully.")

	db.session = session
}

func (db *DB) insert(table string) {
	if err := db.session.Query(`INSERT INTO users (timeline, id, text) VALUES (?, ?, ?)`,
	"me", gocql.TimeUUID(), "DERP").Exec(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func (db *DB) query() {
	if err := db.session.Query(`INSERT INTO users (timeline, id, text) VALUES (?, ?, ?)`,
	"me", gocql.TimeUUID(), "DERP").Exec(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

// defer this
func (db *DB) cleanUp() {
	db.session.Close()
}
