package main

const db_host = "cassandra"
const db_keyspace = "badgerboi"

func main() {
	// time.Sleep(20 * time.Second) // try sleep until cassandra is up

	// connect to db
	db := DB{};
	db.connect(db_host, db_keyspace)
	defer db.cleanUp()

	db.query()
}
