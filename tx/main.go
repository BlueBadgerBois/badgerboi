package main

import (
	"log"
  "net/http"

  "fmt"
  "net"
  "bufio"
  "strings"
)

const db_host = "cassandra"
const db_keyspace = "badgerboi"

func addHandler(w http.ResponseWriter, r *http.Request ) {

}

func main() {
	// connect to db
	db := DB{};
	db.connectWithRetries(db_host, db_keyspace)
	defer db.cleanUp()

	db.initSchema()

	log.Println("running tx server")
	in, _ := net.Listen("tcp", ":8082")
  conn, _ := in.Accept()

  // will listen for message to process ending in newline (\n)
  message, _ := bufio.NewReader(conn).ReadString('\n')
  // output message received
  fmt.Print("Message Received:", string(message))
  // sample process for string received
  newmessage := strings.ToUpper(message)
  // send new string back to client
  conn.Write([]byte(newmessage + "\n"))
}
