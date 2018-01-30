package main

import (
	"bytes"
	"bufio"
	"errors"
	"html/template"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"strconv"
)

const MAX_TRANSACTION_VALIDITY_SECS = 60

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
		"Money: " + centsToDollarsString(user.CurrentMoney))
	}
}

func (handler *Handler) add(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		amount := r.Form.Get("amount")

		amountInCents := stringMoneyToCents(amount)

		user := UserFromUsernameOrCreate(db, username)

		newAmount := user.CurrentMoney + amountInCents

		user.CurrentMoney = newAmount
		db.conn.Save(&user)

		logAddCommand(user)
		logAccountTransaction(user, "add")

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

		user := UserFromUsernameOrCreate(db, responseMap["username"])

		logQuoteCommand(responseMap, user)

		fmt.Fprintf(w, "Success!\n\n Quote Server Response: %v", responseMap)
	}
}

func (handler *Handler) dumplog(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		outfile := r.Form.Get("outfile")

		u := LogItem{Username: username}

		var log_items []LogItem

		if username == "admin" {
			db.conn.Find(&log_items)
		} else {
			db.conn.Where(&u).Find(&log_items)
		}

		xmlLog := writeLogsToFile(outfile, log_items)

		fmt.Fprintf(w, 
			"Username: " + username + "\n" +
			"Outfile: " + outfile + "\n" +
			"Logfile contents: ")
			fmt.Fprintf(w, xmlLog)
	}
}

// Save an ErrorEventLogItem
func logErrorEvent(params map[string]string, user *User) {
	errorEventLogItem := BuildErrorEventLogItemStruct()
	errorEventLogItem.Command = params["command"]
	errorEventLogItem.Username = user.Username
	errorEventLogItem.StockSymbol = params["stockSymbol"]
	errorEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	errorEventLogItem.ErrorMessage = params["errorMessage"]
	username := user.Username

	errorEventLogItem.SaveRecord(username)
}

// Save a UserCommandLogItem for an ADD command
func logAddCommand(user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "ADD"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = ""
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

func logAccountTransaction(user *User, action string) {
	transactionLogItem := BuildAccountTransactionLogItemStruct()
	transactionLogItem.Action = action
	transactionLogItem.Username = user.Username
	transactionLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	transactionLogItem.SaveRecord(username)
}

func logSystemEvent(user *User, params map[string]string) {
	systemEventLogItem := BuildSystemEventLogItemStruct()
	systemEventLogItem.Command = params["command"]
	systemEventLogItem.StockSymbol = params["stockSymbol"]
	systemEventLogItem.Filename = params["filename"]
	systemEventLogItem.Username = user.Username
	systemEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	systemEventLogItem.SaveRecord(username)
}

func logSummaryCommand(user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "DISPLAY_SUMMARY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = "" // No stock symbol for a summary
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
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
	quoteLogItem := BuildQuoteServerLogItemStruct()
	quoteLogItem.Price = params["price"]
	quoteLogItem.StockSymbol = params["stockSymbol"]
	quoteLogItem.Username = params["username"]
	quoteLogItem.QuoteServerTime = params["quoteServerTime"]
	quoteLogItem.Cryptokey = params["cryptokey"]
	username := params["username"]

	quoteLogItem.SaveRecord(username)
}

// Save a UserCommandLogItem for a QUOTE command
func logQuoteCommand(params map[string]string, user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "QUOTE"
	commandLogItem.Username = params["username"]
	commandLogItem.StockSymbol = params["stockSymbol"]
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := params["username"]

	commandLogItem.SaveRecord(username)
}

func quoteResponseToMap(message string) map[string]string {
	splitMessage := strings.Split(message, ",")

	log.Println("Quote server response: ", splitMessage)

	outputMap := map[string]string {
		"price": strings.TrimSpace(splitMessage[0]),
		"username": strings.TrimSpace(splitMessage[1]),
		"stockSymbol": strings.TrimSpace(splitMessage[2]),
		"quoteServerTime": strings.TrimSpace(splitMessage[3]),
		"cryptokey": strings.TrimSpace(splitMessage[4]),
	}

	return outputMap
}

func stringMoneyToCents(amount string) uint { // this needs to be fixed to handle improper inputs (no decimal)
	stringAmountCents, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))
	return uint(stringAmountCents)
}

func authUser(uname string) (User, error) {
	u := User{Username: uname}
	var user User
	if db.conn.First(&user, &u).RecordNotFound() {
		return user, errors.New("User not found!")
	}
	return user, nil
}

/*
* ConvertMoneyToStock:
* Inputs: 
*   - Amount of money you are spending
*   - The price of the stock
* Outputs: 
*   - The number of units of stock you can buy 
*   - The leftover money from the transaction
*/
func convertMoneyToStock(money uint, stockPrice uint) (uint, uint) {
	amnt := (money/stockPrice)
	leftover := (money%stockPrice)
	return amnt, leftover
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

func anyStocksCanBeBought(amountToBuyInCents uint, quotedPriceInCents uint) bool {
	numStocksToBuy, _ := convertMoneyToStock(amountToBuyInCents, quotedPriceInCents)
	return numStocksToBuy > 0
}

func writeLogsToFile(outfile string, log_items []LogItem) string{
	var logFileXML bytes.Buffer 
	logFileXML.WriteString("<?xml version=\"1.0\"?>\n")
	logFileXML.WriteString("<log>\n")

	for i := 0; i < len(log_items); i++ {
		fmt.Println(log_items[i].Data)

		logItem := strings.Replace(log_items[i].Data, "\"", "", -1)
		logItem = strings.Replace(logItem, "}", "", -1)
		logItem = strings.Replace(logItem, "{", "", -1)

		keyValues := strings.Split(logItem, ",")

		logType := strings.Split(keyValues[0], ":")

		logFileXML.WriteString("\t<" + logType[1] + ">\n")

		for i := 0; i < len(keyValues); i++ {
			attribute := strings.Split(keyValues[i], ":")
			logFileXML.WriteString("\t\t<" + attribute[0] + ">")
			logFileXML.WriteString(attribute[1])
			logFileXML.WriteString("</" + attribute[0] + ">\n")
		}
		logFileXML.WriteString("\t</" + logType[1] + ">\n")
	}
	logFileXML.WriteString("</log>\n")

	fmt.Println(logFileXML.String())
	return logFileXML.String()
	//ioutil.WriteFile(outfile, logFileXML.String())
}
