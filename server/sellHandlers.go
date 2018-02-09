package main

import (
	"app/db"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

func (handler *Handler) sell(w http.ResponseWriter, r *http.Request) {
	// Sell the minimum number of stocks needed to make the revenue given in sellAmount. 
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		sellAmount := r.Form.Get("sellAmount") // the amount of money we want to get out of this
		stockSymbol := r.Form.Get("stockSymbol")

		amountToSellInCents := stringMoneyToCents(sellAmount)

		user := db.UserFromUsernameOrCreate(dbw, username)

		logSellCommand(stockSymbol, user)

		quoteResponseMap := getQuoteFromServer(username, stockSymbol)

		err, _ := createSellTransaction(user, stockSymbol, amountToSellInCents,
		stringMoneyToCents(quoteResponseMap["price"]))

		if err != nil {
			fmt.Fprintf(w, "Error!" + err.Error())
			return
		}

		fmt.Fprintf(w,
		"Transaction created. Pending sell transaction:\n\n" +
		"User ID: " + username + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Amount to sell: " + centsToDollarsString(amountToSellInCents) +
		"\nQuoted price: " + quoteResponseMap["price"])
	}
}

func (handler *Handler) commitSell(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")

		user, err := db.UserFromUsername(dbw, username)
		if err != nil {
			fmt.Fprintf(w, "Failure! user does not exist!\n\n")
			errorEventParams := map[string]string {
				"command": "COMMIT_SELL",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)
			return
		}

		transactionToCommit, err := db.NewestPendingTransactionForUser(dbw, user, "sell")
		if err != nil { // If no transaction can be found
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "COMMIT_SELL",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)
			return
		}

		logCommitSellCommand(transactionToCommit.StockSymbol, user)

		// Check to make sure the transaction is not stale
		// later we need to make the job server just update the quote price
		currentTime := dbw.GetCurrentTime()
		timeDifference := currentTime.Sub(transactionToCommit.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			errMsg :=  "Failure! Most recent sell transaction is more than 60 seconds old."
			fmt.Fprintf(w, errMsg)
			errorEventParams := map[string]string {
				"command": "COMMIT_SELL",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(errorEventParams, user)
			return
		}

		numStocksNeeded := stocksNeededToGetAmountInCents(transactionToCommit.AmountInCents,
		transactionToCommit.QuotedStockPrice)

		// tx := dbw.Conn.Begin()
		tx := dbw.Begin()
		tx.Conn.Where(&db.User{Username: user.Username}).First(&user) // reload user
		userHolding, _ := db.BuildStockHolding(tx, user, transactionToCommit.StockSymbol) // grab the user's current holding for the stock

		// Check if user still has enough stocks to make the sale
		if !userHolding.Sufficient(numStocksNeeded) {
			tx.Rollback() // rollback immediately

			errMsg := "Failure! User does not have enough stocks\n\n" +
			"User ID: " + user.Username + "\n" +
			"Stock Symbol: " + transactionToCommit.StockSymbol + "\n" +
			"Number to sell: " + fmt.Sprint(numStocksNeeded) +
			"\nCurrent holdings: " + fmt.Sprint(userHolding.Number)
			errorEventParams := map[string]string {
				"command": "COMMIT_SELL",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(errorEventParams, user)
			fmt.Fprintf(w, errMsg)
			return
		}

		// withdraw the stocks
		userHolding.Withdraw(tx, numStocksNeeded)

		centsToDeposit := moneyInCentsForStocks(numStocksNeeded, transactionToCommit.QuotedStockPrice)

		// give user the money
		user.DepositMoney(tx, centsToDeposit)

		tx.Commit()

		fmt.Fprintf(w, "SELL committed.\n " +
		"symbol: " + transactionToCommit.StockSymbol +
		"\nnum sold: " + fmt.Sprint(numStocksNeeded) +
		"\nquoted price: " + centsToDollarsString(transactionToCommit.QuotedStockPrice) +
		"\namount of money deposited: " + centsToDollarsString(centsToDeposit))
	}
}

func (handler *Handler) cancelSell(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")

		user, err := db.UserFromUsername(dbw, username)
		if err != nil {
			fmt.Fprintf(w, "Failure! user does not exist!\n\n")
			errorEventParams := map[string]string {
				"command": "COMMIT_SELL",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)
			return
		}

		transactionToCancel, err := db.NewestPendingTransactionForUser(dbw, user, "sell")
		if err != nil {
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "CANCEL_SELL",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		logCancelSellCommand(transactionToCancel.StockSymbol, user)

		// later we need to make the job server just update the quote price
		currentTime := dbw.GetCurrentTime()
		timeDifference := currentTime.Sub(transactionToCancel.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			errMsg := "Failure! Most recent SELL transaction is more than 60 seconds old."
			fmt.Fprintf(w, errMsg)
			errorEventParams := map[string]string {
				"command": "CANCEL_SELL",
				"stockSymbol": transactionToCancel.StockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		transactionToCancel.Cancel(dbw)

		successMsg := "SELL cancelled.\n" +
		"symbol: " + transactionToCancel.StockSymbol +
		"\nmoney that was to be made: " + centsToDollarsString(transactionToCancel.AmountInCents)
		fmt.Fprintf(w, successMsg)
	}
}

func createSellTransaction(user *db.User, stockSymbol string, amountToSellInCents uint, quotedPriceInCents uint) (error, *db.Transaction) {
	numStocksNeeded := stocksNeededToGetAmountInCents(amountToSellInCents, quotedPriceInCents)

	// create a transaction record
	sellTransaction := db.BuildSellTransaction(user)
	sellTransaction.StockSymbol = stockSymbol
	sellTransaction.AmountInCents = amountToSellInCents
	sellTransaction.QuotedStockPrice = quotedPriceInCents
	dbw.Conn.NewRecord(sellTransaction)

	userHolding, _ := db.BuildStockHolding(dbw, user, sellTransaction.StockSymbol)

	// Check if user has enough stocks to make the given amount of revenue
	if !userHolding.Sufficient(numStocksNeeded) {
		errMsg := "Failure! User does not have enough stocks\n\n" +
		"User ID: " + user.Username + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Number to sell: " + fmt.Sprint(numStocksNeeded) +
		"\nCurrent holdings: " + fmt.Sprint(userHolding.Number)
		errorEventParams := map[string]string {
			"command": "SELL",
			"stockSymbol": stockSymbol,
			"errorMessage": errMsg,
		}
		logErrorEvent(errorEventParams, user)

		return errors.New(errMsg), sellTransaction
	}
	dbw.Conn.Save(sellTransaction)
	return nil, sellTransaction
}

func stocksNeededToGetAmountInCents(amountToSellInCents uint, stockPriceInCents uint) uint {
	numberToSell := math.Ceil(float64(amountToSellInCents) / float64(stockPriceInCents))
	return uint(numberToSell)
}

func moneyInCentsForStocks(numStocks uint, stockPrice uint) uint {
	return numStocks * stockPrice
}


// Save a UserCommandLogItem for a SELL command
func logSellCommand(stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "SELL"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for a COMMIT_SELL command
func logCommitSellCommand(stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "COMMIT_SELL"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for a COMMIT_SELL command
func logCancelSellCommand(stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "CANCEL_SELL"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}
