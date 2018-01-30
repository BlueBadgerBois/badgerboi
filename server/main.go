package main

import (
	"log"
	"net/http"
	"os"
	"github.com/kabukky/httpscerts"
)

var db = &DB{} // this is global so everything can see it
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
	runJobServer()
}

func runAsWebServer() {
	// generateCertsIfNotPresent()

	// web server handlers
	log.Println("Running web server.")
	http.HandleFunc("/", handler.index)
	http.HandleFunc("/quote", handler.quote)
	http.HandleFunc("/add", handler.add)
	http.HandleFunc("/buy", handler.buy)
	http.HandleFunc("/commitBuy", handler.commitBuy)
	http.HandleFunc("/cancelBuy", handler.cancelBuy)
	http.HandleFunc("/sell", handler.sell)
	http.HandleFunc("/commitSell", handler.commitSell)
	http.HandleFunc("/cancelSell", handler.cancelSell)
	http.HandleFunc("/summary", handler.summary)
	http.HandleFunc("/setBuyAmount", handler.setBuyAmount)
	http.HandleFunc("/cancelSetBuy", handler.cancelSetBuy)
	http.HandleFunc("/setBuyTrigger", handler.setBuyTrigger)
	http.HandleFunc("/setSellAmount", handler.setSellAmount)
	http.HandleFunc("/cancelSetSell", handler.cancelSetSell)
	http.HandleFunc("/setSellTrigger", handler.setSellTrigger)
	http.HandleFunc("/dumplog", handler.dumplog)
	http.ListenAndServe(":8082", nil)
	// http.ListenAndServeTLS(":8082", "cert.pem", "key.pem", nil)
}

func generateCertsIfNotPresent() {
	// Check if the cert files are available.
	err := httpscerts.Check("cert.pem", "key.pem")

	// If they are not available, generate new ones.
	if err != nil {
		err = httpscerts.Generate("cert.pem", "key.pem", "web:8081")
		if err != nil {
			log.Fatal("Error: Couldn't create https certs.")
		}
	}
}

func getServerRole() string {
	role := os.Getenv("ROLE")
	return role
}
