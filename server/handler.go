package main

import (
	"bufio"
	"html/template"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"strconv"
)

type Handler struct {}

func (handler *Handler) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		username := r.Form.Get("username")

		u := User{Username: username}

		var user User
		if db.conn.First(&user, &u).RecordNotFound() {
			fmt.Fprintf(w, "This user doesn't exist!")
			return
		}

		fmt.Fprintf(w,
			"Summary:\n\n" +
			"Username: " + username + "\n" +
			"Money: " + strconv.Itoa(int(user.CurrentMoney)))
	}
}

func (handler *Handler) add(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		amount := r.Form.Get("amount")

		amountFormatted := stringMoneyToCents(amount)

		u := User{Username: username}

		var user User
		db.conn.FirstOrCreate(&user, &u)

		newAmount := user.CurrentMoney + amountFormatted

		user.CurrentMoney = newAmount
		// user.CurrentMoney = 123
		db.conn.Save(&user)

		fmt.Fprintf(w,
		"Success!\n\n" +
		"User ID: " + username + "\n" +
		"Amount Added: " + string(amount))
	}
}

func (handler *Handler) index(w http.ResponseWriter, r *http.Request) {
	temp, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Println("template parsing error: ", err)
	}

	err = temp.Execute(w, nil)
	if err != nil{
		log.Println("template execution error: ", err)
	}
}

func (handler *Handler) quote(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stocksym := r.Form.Get("stocksym")

		log.Println("UID: " + username + ", Stock: " + stocksym)

		//hit quote server
		conn, err := net.Dial("tcp", "quoteServer:3333")

		if err != nil{
			log.Println("error hitting quote server: ", err)
		}

		fmt.Fprintf(conn, username + ", " + stocksym)
		message, _ := bufio.NewReader(conn).ReadString('\n')

		fmt.Fprintf(w,
		"Success!\n\n" +
		"Quote Server Response: " + message)

		log.Println("JSON from message map: ", jsonFromMap(quoteResponseToMap(message)))

	}
}

func bytesToString(bytes []byte) string {
	return string(bytes[:])
}

func jsonFromMap(inputMap map[string]string) string {
	jsonBytes, err := json.Marshal(inputMap)

	if err != nil {
		log.Fatal("Unable to convert map %s to json", inputMap)
	}

	jsonString := bytesToString(jsonBytes)

	return jsonString
}

func enterQueryLogItem(jsonData string) {
	// TODO
}

func quoteResponseToMap(message string) map[string]string {
	// price stockSymbol, username, quoteServerTime, cryptokey

	splitMessage := strings.Split(message, ",")

	log.Println("split message: ", splitMessage)

	outputMap := map[string]string {
		"price": strings.TrimSpace(splitMessage[0]),
		"username": strings.TrimSpace(splitMessage[1]),
		"stockSymbol": strings.TrimSpace(splitMessage[2]),
		"quoteServerTime": strings.TrimSpace(splitMessage[3]),
		"cryptokey": strings.TrimSpace(splitMessage[4]),
	}

	return outputMap
}

func (handler *Handler) buy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		buyAmount := r.Form.Get("buyAmount")
		stockSymbol := r.Form.Get("stockSymbol")

		// amountFormatted := stringMoneyToCents(buyAmount) // needs fix

		u := User{Username: username}

		var user User
		db.conn.FirstOrCreate(&user, &u)

		// newAmount := user.CurrentMoney + amountFormatted

		user.CurrentMoney = 1211
		db.conn.Save(&user)

		fmt.Fprintf(w,
		"Success!\n\n" +
		"User ID: " + username + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Amount bought: " + string(buyAmount))
	}
}

func stringMoneyToCents(amount string) (uint) { // this needs to be fixed
	formattedAmount, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))
	return uint(formattedAmount)
}
