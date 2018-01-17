package main

import (
	"log"
	"os"
	"time"

	"github.com/gocql/gocql"
)

func main() {
	time.Sleep(20 * time.Second) // try sleep until cassandra is up
	cluster := gocql.NewCluster("cassandra")
	cluster.Keyspace = "example"
	cluster.Consistency = gocql.Quorum

	session, sess_err := cluster.CreateSession()

	if sess_err != nil {
		log.Fatal(sess_err)
		os.Exit(1)
	}

	defer session.Close()

	if err := session.Query(`INSERT INTO tweet (timeline, id, text) VALUES (?, ?, ?)`,
	"me", gocql.TimeUUID(), "hello world").Exec(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
