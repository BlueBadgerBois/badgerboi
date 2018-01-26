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
	"time"
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

		amountToBuyInCents := stringMoneyToCents(buyAmount)

		user := db.userFromUsernameOrCreate(username)

		logBuyCommand(stockSymbol, user)

		quoteResponseMap := getQuoteFromServer(username, stockSymbol)
		err, _ := createBuyTransaction(user, stockSymbol, amountToBuyInCents, quoteResponseMap["price"])
		
		if err != nil {
			fmt.Fprintf(w, "Error!", err)
			return
		}

		fmt.Fprintf(w,
			"Transaction created. Pending buy transaction:\n\n" +
			"User ID: " + username + "\n" +
			"Stock Symbol: " + stockSymbol + "\n" +
			"Amount to buy: " + centsToDollarsString(amountToBuyInCents) +
			"\nQuoted price: " + quoteResponseMap["price"])
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

func createBuyTransaction(user *User, stockSymbol string, amountToBuyInCents uint, quotePrice string) (error, uint) {
	// create a transaction record
	buyTransaction := buildBuyTransaction(user)
	buyTransaction.StockSymbol = stockSymbol
	buyTransaction.AmountInCents = amountToBuyInCents
	buyTransaction.QuotedStockPrice = uint(stringMoneyToCents(quotePrice))
	db.conn.Create(&buyTransaction)

	if !user.HasEnoughMoney(amountToBuyInCents) {
		errorEventParams := map[string]string {
			"command": "BUY",
			"stockSymbol": stockSymbol,
			"errorMessage": "Insufficient funds",
		}
		logErrorEvent(errorEventParams, user)

		return errors.New(
			"Failure! User does not have enough money\n\n" +
			"User ID: " + user.Username + "\n" +
			"Stock Symbol: " + stockSymbol + "\n" +
			"Amount to buy: " + centsToDollarsString(amountToBuyInCents) +
			"\nCurrent funds: " + centsToDollarsString(user.CurrentMoney)), 0
	}
	return nil, buyTransaction.ID
}

// Save a UserCommandLogItem for a BUY command
func logBuyCommand(stockSymbol string, user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

func (handler *Handler) commitBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")


		user, err := db.userFromUsername(username)
		if err != nil {
			fmt.Fprintf(w, "Failure! user does not exist!\n\n")
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		transactionToCommit, err := db.newestTransactionForUser(user)
		if err != nil {
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)

			return
		}


		// error unless transactionToCommit exists



		logCommitBuyCommand(transactionToCommit.StockSymbol, user)

		// check if the current time is more than 60s after the created_at time of the transaction
		log.Println("Transaction created_at: ", transactionToCommit.CreatedAt)
		currentTime := time.Now()
		log.Println("Current time: ", currentTime)

		if !user.HasEnoughMoney(transactionToCommit.AmountInCents) {
			fmt.Fprintf(w,
			"Failure! User does not have enough money\n\n" +
			"User ID: " + user.Username + "\n" +
			"Stock Symbol: " + transactionToCommit.StockSymbol + "\n" +
			"Amount to buy: " + centsToDollarsString(transactionToCommit.AmountInCents) +
			"\nCurrent funds: " + centsToDollarsString(user.CurrentMoney))

			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": "Insufficient funds",
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		// // create a transaction record
		// buyTransaction := buildBuyTransaction(user)
		// buyTransaction.StockSymbol = stockSymbol
		// buyTransaction.AmountInCents = amountToBuyInCents
		// buyTransaction.QuotedStockPrice = uint(stringMoneyToCents(quoteResponseMap["price"]))
		// db.conn.Create(&buyTransaction)

		fmt.Fprintf(w,
		"Transaction created. Pending buy transaction:\n\n")
		// "User ID: " + username + "\n" +
		// "Stock Symbol: " + stockSymbol + "\n" +
		// "Amount to buy: " + buyAmount +
		// "\nQuoted price: " + quoteResponseMap["price"])
	}
}

// Save a UserCommandLogItem for a COMMIT_BUY command
func logCommitBuyCommand(stockSymbol string, user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "COMMIT_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

// Save an ErrorEventLogItem
func logErrorEvent(params map[string]string, user *User) {
	errorEventLogItem := buildErrorEventLogItemStruct()
	errorEventLogItem.Command = params["command"]
	errorEventLogItem.Username = user.Username
	errorEventLogItem.StockSymbol = params["stockSymbol"]
	errorEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	errorEventLogItem.ErrorMessage = params["errorMessage"]
	username := user.Username

	errorEventLogItem.SaveRecord(username)
}

// we should probably use transactions for all these db operations
// ex.
// tx := db.conn.Begin()
// tx.Find(), tx.FirstOrCreate(), etc.
// if err { tx.Rollback() }
func (handler *Handler) setBuyAmount(w http.ResponseWriter, r *http.Request) {
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

		amountToBuyInCents := stringMoneyToCents(buyAmount)
		thresholdInCents := stringMoneyToCents(threshold)
		quoteResponseMap := getQuoteFromServer(username, stockSymbol)

		transactionErr, transactionID := createBuyTransaction(&user, stockSymbol, amountToBuyInCents, quoteResponseMap["price"])
		if transactionErr != nil {
			fmt.Fprintf(w, "Error: ", transactionErr)
			return
		}

		/*
		* We created the transaction, we also need to remove some money from the user
		* In the requirements it says we should store this in a reserve account,
		* but for now I was thinking just remove the money from the user,
		* then give it back if the buy is cancelled
		*/
		user.CurrentMoney = user.CurrentMoney - uint(amountToBuyInCents)
		db.conn.Save(&user)
		// =================== May want to remove it change above
		
		buyTrigger := buildBuyTrigger(&user)
		buyTrigger.StockSym = stockSymbol
		buyTrigger.Amount = uint(amountToBuyInCents)
		buyTrigger.PriceThreshold = uint(thresholdInCents)
		buyTrigger.TransactionID = transactionID
		db.conn.Create(&buyTrigger)

		fmt.Fprintf(w, 
			"Trigger was successfully created!\n\n" +
			"Amount withdrawn: $" + centsToDollarsString(amountToBuyInCents) +
			"\nThis will be held until...\n" +
			"Stock: " + stockSymbol +
			" reaches $" + centsToDollarsString(thresholdInCents))
	}
}

// Save a UserCommandLogItem for an ADD command
func logAddCommand(user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "ADD"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = ""
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

func logAccountTransaction(user *User, action string) {
	transactionLogItem := buildAccountTransactionLogItemStruct()
	transactionLogItem.Action = action
	transactionLogItem.Username = user.Username
	transactionLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	transactionLogItem.SaveRecord(username)
}

func logSystemEvent(user *User, params map[string]string) {
	systemEventLogItem := buildSystemEventLogItemStruct()
	systemEventLogItem.Command = params["command"]
	systemEventLogItem.StockSymbol = params["stockSymbol"]
	systemEventLogItem.Filename = params["filename"]
	systemEventLogItem.Username = user.Username
	systemEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	systemEventLogItem.SaveRecord(username)
}

func logSummaryCommand(user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
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
	quoteLogItem := buildQuoteServerLogItemStruct()
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
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "QUOTE"
	commandLogItem.Username = params["username"]
	commandLogItem.StockSymbol = params["stockSymbol"]
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := params["username"]

	commandLogItem.SaveRecord(username)
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