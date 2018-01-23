package main

import (
	"log"
)

const db_host = "cassandra"
const db_keyspace = "badgerboi"

var db = DB{} // this is global so everything can see it

func main() {
	// connect to db
	db.init()
	defer db.cleanUp()

	log.Println("Running tx server")
}
