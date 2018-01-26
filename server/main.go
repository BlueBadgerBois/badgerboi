package main

import (
	"log"
	"net/http"
	"os"

	"time" // possibly move to another file

	"github.com/kabukky/httpscerts"
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
	// TODO: Move all this job server stuff to another file
	log.Println("Running job server.")
	tickChan := time.NewTicker(20*time.Second).C
	for {
		<- tickChan
		log.Println("")
		log.Println("Checking all triggers")
		log.Println("")
		checkTriggers()
	}
}

func checkTriggers() {
	var triggers []Trigger
	db.conn.Find(&triggers)
	for _, trig := range triggers {
		var user User
		db.conn.First(&user)
		log.Println("Username: " + user.Username)
		log.Println("StockSymbol: " + trig.StockSym)
		log.Println("Type of trigger: " + trig.Type)
		log.Println("Price threshold: " + centsToDollarsString(trig.PriceThreshold))

		responseMap := getQuoteFromServer(user.Username, trig.StockSym)
		log.Println("Current price: " + responseMap["price"])

		quotePriceStr := stringMoneyToCents(responseMap["price"])
		if trig.Type == "buy" && quotePriceStr < trig.PriceThreshold {
			log.Println("We are below the threshold so we can buy!")
			// we need to commit the buy
			// however, we already have taken the funds out of the users account
			// so it might be hard to reuse commitBuy code...
			// to counteract, we could temporarily add the money back, then call on
			// commitBuy
		} else if trig.Type == "sell" && quotePriceStr > trig.PriceThreshold {
			log.Println("We are above the threshold so we can sell!")
		}

		// if success:
		//   Remove the trigger from database?***
	}
}

func runAsWebServer() {
	// generateCertsIfNotPresent()

	// web server handlers
	log.Println("Running web server.")
	http.HandleFunc("/", handler.index)
	http.HandleFunc("/quote", handler.quote)
	http.HandleFunc("/add", handler.add)
	http.HandleFunc("/buy", handler.buy)
	http.HandleFunc("/summary", handler.summary)
	http.HandleFunc("/commitBuy", handler.commitBuy)
	http.HandleFunc("/setBuyAmount", handler.setBuyAmount)
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
