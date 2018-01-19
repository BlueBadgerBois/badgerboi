package main

const db_host = "cassandra"
const db_keyspace = "badgerboi"

func main() {
	// connect to db
	db := DB{};
	db.connectWithRetries(db_host, db_keyspace)
	defer db.cleanUp()

	db.query()
}
