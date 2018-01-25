package main

import (
	"log"
	"fmt"
	"net"
	"net/http"
	"html/template"
	"strconv"
	"strings"
	"bufio"
)

type Handler struct {}

func (handler *Handler) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		r.ParseForm()
		uID := r.Form.Get("uID")

		u := User{Username: uID}

		var user User
		if db.conn.First(&user, &u).RecordNotFound() {
			fmt.Fprintf(w, "This user doesn't exist!")
			return
		}

		fmt.Fprintf(w,
			"Summary:\n\n" +
			"Username: " + uID + "\n" +
			"Money: " + strconv.Itoa(int(user.CurrentMoney)))
	}
}

func (handler *Handler) add(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		uID := r.Form.Get("uID")
		amount := r.Form.Get("amount")

		amountFormatted := stringMoneyToCents(amount)

		u := User{Username: uID}

		var user User
		db.conn.FirstOrCreate(&user, &u)

		newAmount := user.CurrentMoney + amountFormatted

		user.CurrentMoney = newAmount
		// user.CurrentMoney = 123
		db.conn.Save(&user)

		fmt.Fprintf(w,
		"Success!\n\n" +
		"User ID: " + uID + "\n" +
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
	temp, err := template.ParseFiles("templates/quote.html")
	if err != nil {
		log.Println("template parsing error: ", err)
	}

	err = temp.Execute(w, nil)
	if err != nil{
		log.Println("template execution error: ", err)
	}

	if r.Method == "POST" {
		r.ParseForm()
		uID := r.Form.Get("uID")
		stocksym := r.Form.Get("stocksym")

		log.Println("UID: " + uID + ", Stock: " + stocksym)

		//hit quote server
		conn, err := net.Dial("tcp", "quoteServer:3333")

		if err != nil{
			log.Println("error hitting quote server: ", err)
		}

		fmt.Fprintf(conn, uID + ", " + stocksym)
		message, _ := bufio.NewReader(conn).ReadString('\n')

		fmt.Fprintf(w,
		"Success!\n\n" +
		"Quote Server Response: " + message)
	}
}

func (handler *Handler) buy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		uID := r.Form.Get("uID")
		buyAmount := r.Form.Get("buyAmount")
		stockSymbol := r.Form.Get("stockSymbol")

		// amountFormatted := stringMoneyToCents(buyAmount) // needs fix

		u := User{Username: uID}

		var user User
		db.conn.FirstOrCreate(&user, &u)

		// newAmount := user.CurrentMoney + amountFormatted

		user.CurrentMoney = 1211
		db.conn.Save(&user)

		fmt.Fprintf(w,
		"Success!\n\n" +
		"User ID: " + uID + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Amount bought: " + string(buyAmount))
	}
}

func stringMoneyToCents(amount string) (uint) {
	formattedAmount, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))
	return uint(formattedAmount)
}
