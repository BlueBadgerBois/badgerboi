package main

import (
	"bufio"
	"html/template"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"strconv"
  "errors"
)

type Handler struct {}

func (handler *Handler) summary(w http.ResponseWriter, r *http.Request) {
	// TODO iterate through all holdings. Need to timplement buy first
	if r.Method == "GET" {
		r.ParseForm()
		username := r.Form.Get("username")

		u := User{Username: username}

		var user User
		if db.conn.First(&user, &u).RecordNotFound() {
			fmt.Fprintf(w, "This user doesn't exist!")
			return
		}

		logSummaryCommand(&user)

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

		amountInCents := stringMoneyToCents(amount)

		user := db.userFromUsernameOrCreate(username)

		newAmount := user.CurrentMoney + amountInCents

		user.CurrentMoney = newAmount
		db.conn.Save(&user)

		logAddCommand(user)

		fmt.Fprintf(w,
		"Success!\n\n" +
		"User ID: " + username + "\n" +
		"Amount added in dollars: " + centsToDollarsString(amountInCents) +
		"\nCents: " + strconv.Itoa(int(amountInCents)))
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
		stockSymbol := r.Form.Get("stocksym")

		responseMap := getQuoteFromServer(username, stockSymbol)

		user := db.userFromUsernameOrCreate(responseMap["username"])

		logQuoteCommand(responseMap, user)

		fmt.Fprintf(w, "Success!\n\n Quote Server Response: %v", responseMap)
	}
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

func (handler *Handler) setBuyTrigger(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    username := r.Form.Get("username")
    stockSymbol := r.Form.Get("stockSymbol")
    buyAmount := r.Form.Get("buyAmount")
    threshold := r.Form.Get("threshold")

    user, err := authUser(username)
    if err != nil {
      fmt.Fprintf(w, "Error: ", err)
      return
    }

    buyAmountFormatted, _ := strconv.Atoi(buyAmount)
    thresholdFormatted, _ := strconv.Atoi(threshold)

    t := Trigger{
      UserID: user.ID,
      StockSym: stockSymbol,
      BuyAmount: uint(buyAmountFormatted),
      PriceThreshold: uint(thresholdFormatted),
    }

    db.conn.Create(&t)

    log.Println(stockSymbol, buyAmount, threshold)
  }
}

// Save a UserCommandLogItem for an ADD command
func logAddCommand(user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "ADD"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = ""
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)

	commandLogItem.SaveRecord()
}

func logSummaryCommand(user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "DISPLAY_SUMMARY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = "" // No stock symbol for a summary
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.SaveRecord()
}

// Fetch a quote from the quote server and log it
func getQuoteFromServer(username string, stockSymbol string) map[string]string {
	conn, err := net.Dial("tcp", "quoteServer:3333")

	if err != nil{
		log.Println("error hitting quote server: ", err)
	}

	fmt.Fprintf(conn, username + ", " + stockSymbol)
	message, _ := bufio.NewReader(conn).ReadString('\n')

	responseMap := quoteResponseToMap(message)

	logQuoteServer(responseMap)

	return responseMap
}

// Save a QuoteServerLogItem
func logQuoteServer(params map[string]string) {
	quoteLogItem := buildQuoteServerLogItemStruct()
	quoteLogItem.Price = params["price"]
	quoteLogItem.StockSymbol = params["stockSymbol"]
	quoteLogItem.Username = params["username"]
	quoteLogItem.QuoteServerTime = params["quoteServerTime"]
	quoteLogItem.Cryptokey = params["cryptokey"]

	quoteLogItem.SaveRecord()
}

// Save a UserCommandLogItem for a QUOTE command
func logQuoteCommand(params map[string]string, user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "QUOTE"
	commandLogItem.Username = params["username"]
	commandLogItem.StockSymbol = params["stockSymbol"]
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)

	commandLogItem.SaveRecord()
}

func quoteResponseToMap(message string) map[string]string {
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


func stringMoneyToCents(amount string) (uint) { // this needs to be fixed to handle improper inputs (no decimal)
	formattedAmount, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))
	return uint(formattedAmount)
}

func authUser(uname string) (User, error) {
  u := User{Username: uname}
  var user User
	if db.conn.First(&user, &u).RecordNotFound() {
    return user, errors.New("User not found!")
  }
  return user, nil
}

func centsToCentsString(cents uint) string {
	return strconv.Itoa(int(cents))
}

func centsToDollarsString(cents uint) string {
	// 2 is precision after decimal point to display
	return strconv.FormatFloat(centsToDollars(cents), 'f', 2, 64)
}
func centsToDollars(cents uint) float64 {
	return float64(cents) / 100
}
