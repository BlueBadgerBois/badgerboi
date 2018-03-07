package main

import (
	"app/db"

	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const MAX_TRANSACTION_VALIDITY_SECS = 60
const MAX_QUOTE_VALIDITY_SECS = 60

type Handler struct {}

func (handler *Handler) summary(w http.ResponseWriter, r *http.Request) {
	// TODO iterate through all holdings. Need to timplement buy first

	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")

		u := db.User{Username: username}

		var user db.User
		if dbw.Conn.First(&user, &u).RecordNotFound() {
			fmt.Fprintf(w, "This user doesn't exist!")
			return
		}

		logSummaryCommand(txNum, &user)

		fmt.Fprintf(w,
		"Summary:\n\n" +
		"Username: " + username + "\n" +
		"Money: " + centsToDollarsString(user.CurrentMoney))
	}
}

func (handler *Handler) add(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		amount := r.Form.Get("amount")

		amountInCents := stringMoneyToCents(amount)

		user := db.UserFromUsernameOrCreate(dbw, username)

		newAmount := user.CurrentMoney + amountInCents

		user.CurrentMoney = newAmount
		dbw.Conn.Save(&user)

		logAddCommand(txNum, user)
		logAccountTransaction(txNum, user, "add")

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
	txNum := db.NewTxNum(dbw)
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stocksym")

		responseMap := getQuoteWithCaching(txNum, username, stockSymbol)

		user := db.UserFromUsernameOrCreate(dbw, responseMap["username"])

		logQuoteCommand(txNum, responseMap, user)

		fmt.Fprintf(w, "Success!\n\n Quote Server Response: %v", responseMap)
	}
}

func (handler *Handler) dumplog(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		outfile := r.Form.Get("outfile")

		u := db.LogItem{Username: username}

		var log_items []db.LogItem

		if username == "admin" {
			dbw.Conn.Find(&log_items)
		} else {
			dbw.Conn.Where(&u).Find(&log_items)
		}

		xmlLog := writeLogsToFile(outfile, log_items)
		w.Header().Set("Content-Description", "File Transfer")
		w.Header().Set("Content-Transfer-Encoding", "binary")
		w.Header().Set("Content-Disposition", "attachment; filename=" + outfile)
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeFile(w, r, "./" + outfile)

		fmt.Fprintf(w,
			"Username: " + username + "\n" +
			"Outfile: " + outfile + "\n" +
			"Logfile contents:\n")
			fmt.Fprintf(w, xmlLog)
	}
}

// Save an ErrorEventLogItem
func logErrorEvent(txNum uint, params map[string]string, user *db.User) {
	errorEventLogItem := db.BuildErrorEventLogItemStruct()
	errorEventLogItem.Command = params["command"]
	errorEventLogItem.Username = user.Username
	errorEventLogItem.StockSymbol = params["stockSymbol"]
	errorEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	errorEventLogItem.ErrorMessage = params["errorMessage"]
	errorEventLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	errorEventLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for an ADD command
func logAddCommand(txNum uint, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "ADD"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = ""
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func logAccountTransaction(txNum uint, user *db.User, action string) {
	transactionLogItem := db.BuildAccountTransactionLogItemStruct()
	transactionLogItem.Action = action
	transactionLogItem.Username = user.Username
	transactionLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	transactionLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	transactionLogItem.SaveRecord(dbw, username)
}

func logSystemEvent(txNum uint, user *db.User, params map[string]string) {
	systemEventLogItem := db.BuildSystemEventLogItemStruct()
	systemEventLogItem.Command = params["command"]
	systemEventLogItem.StockSymbol = params["stockSymbol"]
	systemEventLogItem.Filename = params["filename"]
	systemEventLogItem.Username = user.Username
	systemEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	systemEventLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	systemEventLogItem.SaveRecord(dbw, username)
}

func logSummaryCommand(txNum uint, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "DISPLAY_SUMMARY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = "" // No stock symbol for a summary
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func writeQuoteToCache(symbol string, quote string) {
	err := cacheClient.SetKeyWithExpirationInSecs(symbol, quote, MAX_QUOTE_VALIDITY_SECS)
	if err != nil {
		fmt.Println("Error caching quote. Symbol: ", symbol, " Quote: ", quote, "error: ", err)
	}
}

func quoteServerUrl() string {
	url := os.Getenv("QUOTE_SERVER_URL")
	return url
}

// Fetch a quote from the quote server and log it
func getQuoteWithCaching(txNum uint, username string, stockSymbol string) map[string]string {
	responseMap := map[string]string {}

	quoteFromCache, err := cacheClient.GetKeyWithStringVal(stockSymbol)

	if err == nil { // Quote was found in the cache
		fmt.Println("Key ", stockSymbol, " found in cache")
		responseMap["price"] = quoteFromCache
		responseMap["stockSymbol"] = stockSymbol
		responseMap["username"] = username
	} else {
		responseMap = getQuoteFromServer(txNum, username, stockSymbol)

		fmt.Println("Writing ", stockSymbol, "to cache")
		writeQuoteToCache(stockSymbol, responseMap["price"])
	}

	return responseMap
}

// Fetch a quote from the quote server and log it
func getQuoteFromServer(txNum uint, username string, stockSymbol string) map[string]string {
	responseMap := map[string]string {}

	conn, err := net.Dial("tcp", quoteServerUrl())
	if err != nil{
		log.Println("error hitting quote server: ", err)
	}

	fmt.Fprintf(conn, stockSymbol + "," + username + "\n")
	message, _ := bufio.NewReader(conn).ReadString('\n')
	responseMap = quoteServerResponseToMap(message)

	logQuoteServer(txNum, responseMap)

	return responseMap
}

// Save a QuoteServerLogItem
func logQuoteServer(txNum uint, params map[string]string) {
	quoteLogItem := db.BuildQuoteServerLogItemStruct()
	quoteLogItem.Price = params["price"]
	quoteLogItem.StockSymbol = params["stockSymbol"]
	quoteLogItem.Username = params["username"]
	quoteLogItem.QuoteServerTime = params["quoteServerTime"]
	quoteLogItem.Cryptokey = params["cryptokey"]
	quoteLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := params["username"]

	quoteLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for a QUOTE command
func logQuoteCommand(txNum uint, params map[string]string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "QUOTE"
	commandLogItem.Username = params["username"]
	commandLogItem.StockSymbol = params["stockSymbol"]
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := params["username"]

	commandLogItem.SaveRecord(dbw, username)
}

func quoteServerResponseToMap(message string) map[string]string {
	splitMessage := strings.Split(message, ",")

	outputMap := map[string]string {
		"price": strings.TrimSpace(splitMessage[0]),
		"stockSymbol": strings.TrimSpace(splitMessage[1]),
		"username": strings.TrimSpace(splitMessage[2]),
		"quoteServerTime": strings.TrimSpace(splitMessage[3]),
		"cryptokey": strings.TrimSpace(splitMessage[4]),
	}

	return outputMap
}

func stringMoneyToCents(amount string) uint { // this needs to be fixed to handle improper inputs (no decimal)
	stringAmountCents, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))
	return uint(stringAmountCents)
}

func authUser(uname string) (db.User, error) {
	u := db.User{Username: uname}
	var user db.User
	if dbw.Conn.First(&user, &u).RecordNotFound() {
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
	if stockPrice == 0 {
		return 0, money
	}
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

func writeLogsToFile(outfile string, log_items []db.LogItem) string{
	var logFileXML bytes.Buffer 
	logFileXML.WriteString("<?xml version=\"1.0\"?>\n")
	logFileXML.WriteString("<log>\n")

	for i := 0; i < len(log_items); i++ {
		//fmt.Println(log_items[i].Data)

		logItem := strings.Replace(log_items[i].Data, "\"", "", -1)
		logItem = strings.Replace(logItem, "}", "", -1)
		logItem = strings.Replace(logItem, "{", "", -1)

		keyValues := strings.Split(logItem, ",")

		logType := strings.Split(keyValues[0], ":")

		logFileXML.WriteString("\t<" + logType[1] + ">\n")

		logFileXML.WriteString("\t\t<timestamp>")
		logFileXML.WriteString(strconv.FormatInt(log_items[i].CreatedAt.UnixNano()/1000000, 10))
		logFileXML.WriteString("</timestamp>\n")

		for i := 1; i < len(keyValues); i++ {
			attribute := strings.Split(keyValues[i], ":")
			logFileXML.WriteString("\t\t<" + attribute[0] + ">")
			logFileXML.WriteString(attribute[1])
			logFileXML.WriteString("</" + attribute[0] + ">\n")
		}
		logFileXML.WriteString("\t</" + logType[1] + ">\n")
	}
	logFileXML.WriteString("</log>\n")

	//fmt.Println(logFileXML.String())
	err := ioutil.WriteFile(outfile, []byte(logFileXML.String()), 0644)
	if err != nil {
		log.Println("Error writing to file " + outfile)
		panic(err)
	}

	return logFileXML.String()
}
