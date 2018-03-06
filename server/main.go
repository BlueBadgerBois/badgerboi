package main

import (
	"app/db"
	"app/cache"
	"fmt"
	"log"
	"net/http"
	"os"
	"github.com/kabukky/httpscerts"
)

var dbw = &db.DBW{} // this is global so everything can see it
var cacheClient = &cache.Cache{} // this is global so everything can see it
var handler = Handler{}
// var currentTxNum int
type fn func(http.ResponseWriter, *http.Request)

const WEB_ROLE = "web"
const JOB_ROLE = "job"

func main() {
	//.Connect to db
	dbw.Init()
	cacheClient.Init()
	defer dbw.CleanUp()

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
	_, err := cacheClient.Client.Ping().Result()
	if err != nil {
		fmt.Println("**************************************")
		fmt.Println("!!!!! Failed connecting to redis !!!!!")
		fmt.Println("**************************************")
	} else {
		fmt.Println("Connected to redis")
	}

	cacheClient.GetCurrUnixTimeInSecs()

	// web server handlers
	log.Println("Running web server.")
	http.HandleFunc("/", handler.index)
	http.Handle("/quote", newTransaction(handler.quote))
	http.Handle("/add", newTransaction(handler.add))
	http.Handle("/buy", newTransaction(handler.buy))
	http.Handle("/commitBuy", newTransaction(handler.commitBuy))
	http.Handle("/cancelBuy", newTransaction(handler.cancelBuy))
	http.Handle("/sell", newTransaction(handler.sell))
	http.Handle("/commitSell", newTransaction(handler.commitSell))
	http.Handle("/cancelSell", newTransaction(handler.cancelSell))
	http.Handle("/summary", newTransaction(handler.summary))
	http.Handle("/setBuyAmount", newTransaction(handler.setBuyAmount))
	http.Handle("/cancelSetBuy", newTransaction(handler.cancelSetBuy))
	http.Handle("/setBuyTrigger", newTransaction(handler.setBuyTrigger))
	http.Handle("/setSellAmount", newTransaction(handler.setSellAmount))
	http.Handle("/cancelSetSell", newTransaction(handler.cancelSetSell))
	http.Handle("/setSellTrigger", newTransaction(handler.setSellTrigger))
	http.Handle("/dumplog", newTransaction(handler.dumplog))
	http.ListenAndServe(":8082", nil)
	// http.ListenAndServeTLS(":8082", "cert.pem", "key.pem", nil)
}

//Middleware
func newTransaction(next fn) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		nextHandler := http.HandlerFunc(next)
		// txNum := db.NewTxNum(dbw)
		// currentTxNum = txNum
		nextHandler.ServeHTTP(w, r)
	})
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
