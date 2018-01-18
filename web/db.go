package main

import (
	"log"
	"os"
	"github.com/gocql/gocql"
)

type DB struct {
	host string
	cluster *gocql.ClusterConfig
	session *gocql.Session
}

func (db *DB) connect(host string, keyspace string) {
	db.host = host
	db.cluster = gocql.NewCluster(db.host)
	db.cluster.Keyspace = keyspace
	db.cluster.Consistency = gocql.Quorum

	session, sess_err := db.cluster.CreateSession()

	if sess_err != nil {
		log.Fatal(sess_err)
		os.Exit(1)
	}

	log.Println("looks like we connected to the db")

	db.session = session
}

// func (db *DB) insert(table string) {
// 	if err := db.session.Query(`INSERT INTO users (timeline, id, text) VALUES (?, ?, ?)`,
// 	"me", gocql.TimeUUID(), "DERP").Exec(); err != nil {
// 		log.Fatal(err)
// 		os.Exit(1)
// 	}
// }

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
