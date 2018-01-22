package main

// import (
// 	"log"
// )

const db_host = "cassandra"
const db_keyspace = "badgerboi"


func main() {
	// connect to db
	db := DB{}
	db.init()
	defer db.cleanUp()

	// db.conn.Create(&User{Username: "chris"})

	log.Println("Running tx server")
}
