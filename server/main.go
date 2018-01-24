package main

import (
	"log"
	"net/http"
	"os"
)

var db = DB{} // this is global so everything can see it
var handler = Handler{}

const WEB_ROLE = "web"
const JOB_ROLE = "job"

func main() {
	// connect to db
	db.init()
	defer db.cleanUp()

	serverRole := getServerRole();

	switch serverRole {
	case JOB_ROLE:
		runAsJobServer()
	case WEB_ROLE:
		runAsWebServer()
	default:
		log.Println("Unknown server role specified in environment: ", serverRole)
	}
}

func runAsJobServer() {
	log.Println("Running job server.")
}

func runAsWebServer() {
	// web server handlers
	log.Println("Running web server.")
	http.HandleFunc("/", handler.index)
	http.HandleFunc("/quote", handler.quote)
	http.HandleFunc("/add", handler.add)
	http.HandleFunc("/buy", handler.buy)
	http.ListenAndServe(":8082", nil)
}

func getServerRole() string {
	role := os.Getenv("ROLE")
	return role
}
