package main

import (
	"log"
  "net/http"
)

const db_host = "cassandra"
const db_keyspace = "badgerboi"

var db = DB{} // this is global so everything can see it
var handler = Handler{}

func main() {
	// connect to db
	db.init()
	defer db.cleanUp()

  http.HandleFunc("/", handler.index)
  http.ListenAndServe(":8082", nil)

	log.Println("Running tx server")
}
