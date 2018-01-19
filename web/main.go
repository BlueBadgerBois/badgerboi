package main

import (
	"fmt"
	"net/http"
)

const db_host = "cassandra"
const db_keyspace = "badgerboi"

func handler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, " Hello %s", r.URL.Path[1:])
}

func main() {
	// time.Sleep(20 * time.Second) // try sleep until cassandra is up

	http.HandleFunc("/", handler)
	http.ListenAndServe("0.0.0.0:8080", nil)

	// connect to db
	db := DB{};
	db.connect(db_host, db_keyspace)
	defer db.cleanUp()

	db.query()
}
