package main

import (
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

	commandLogItem.SaveRecord()
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

	commandLogItem.SaveRecord()
}

// Save an ErrorEventLogItem
func logErrorEvent(params map[string]string, user *User) {
	errorEventLogItem := buildErrorEventLogItemStruct()
	errorEventLogItem.Command = params["command"]
	errorEventLogItem.Username = user.Username
	errorEventLogItem.StockSymbol = params["stockSymbol"]
	errorEventLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	errorEventLogItem.ErrorMessage = params["errorMessage"]

	errorEventLogItem.SaveRecord()
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

		// update user's money, this is now stored in the trigger
		user.CurrentMoney = user.CurrentMoney - uint(amountToBuyInCents)
		db.conn.Save(&user)
		
		buyTrigger := buildBuyTrigger(&user)
		buyTrigger.StockSym = stockSymbol
		buyTrigger.Amount = uint(amountToBuyInCents)
		buyTrigger.PriceThreshold = uint(thresholdInCents)
		db.conn.Create(&buyTrigger)

		fmt.Fprintf(w, 
			"Trigger was successfully created!\n\n" +
			"Amount withdrawn: $" + centsToDollarsString(amountToBuyInCents) +
			"\nThis will be held until...\n" +
			"Stock: " + stockSymbol +
			" reaches $" + centsToDollarsString(thresholdInCents))
	}
}

func (handler *Handler) cancelSetBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")

		user, err := authUser(username)
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		t := Trigger{UserID: user.ID, StockSym: stockSymbol}
		var trig Trigger;
		db.conn.First(&trig, &t)

		user.CurrentMoney = user.CurrentMoney + trig.Amount

		// ======= START TRANSACTION =======
		tx := db.conn.Begin()

		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		if err := tx.Delete(&trig).Error; err != nil {
			tx.Rollback()
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		tx.Commit()
		// ======== END TRANSACTION ========
	}
}

func (handler *Handler) setSellAmount(w http.ResponseWriter, r *http.Request) {
	return
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
	amntToBuy := (money/stockPrice)
	leftover := (money%stockPrice)
	return amntToBuy,leftover
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
