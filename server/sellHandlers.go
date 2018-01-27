package main

import (
	"errors"
	"fmt"
	"math"
	"net/http"
)

func stocksNeededToGetAmountInCents(amountToSellInCents uint, stockPriceInCents uint) uint {
	numberToSell := math.Ceil(float64(amountToSellInCents) / float64(stockPriceInCents))
	return uint(numberToSell)
}

func (handler *Handler) sell(w http.ResponseWriter, r *http.Request) {
	// Sell the minimum number of stocks needed to make the revenue given in sellAmount. 
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		sellAmount := r.Form.Get("sellAmount") // the amount of money we want to get out of this
		stockSymbol := r.Form.Get("stockSymbol")

		amountToSellInCents := stringMoneyToCents(sellAmount)

		user := db.userFromUsernameOrCreate(username)

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

func createSellTransaction(user *User, stockSymbol string, amountToSellInCents uint, quotedPriceInCents uint) (error, *Transaction) {
	numStocksNeeded := stocksNeededToGetAmountInCents(amountToSellInCents, quotedPriceInCents)

	// create a transaction record
	sellTransaction := buildSellTransaction(user)
	sellTransaction.StockSymbol = stockSymbol
	sellTransaction.AmountInCents = amountToSellInCents
	sellTransaction.QuotedStockPrice = quotedPriceInCents
	db.conn.NewRecord(sellTransaction)

	userHolding, _ := db.stockHolding(user, sellTransaction.StockSymbol)

	// Check if user has enough stocks to make the given amount of revenue
	if !userHolding.sufficient(numStocksNeeded) {
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
	db.conn.Save(sellTransaction)
	return nil, sellTransaction
}

// Save a UserCommandLogItem for a SELL command
func logSellCommand(stockSymbol string, user *User) {
	commandLogItem := buildUserCommandLogItemStruct()
	commandLogItem.Command = "SELL"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)

	commandLogItem.SaveRecord()
}

// // Save a UserCommandLogItem for a COMMIT_SELL command
// func logCommitSellCommand(stockSymbol string, user *User) {
// 	commandLogItem := buildUserCommandLogItemStruct()
// 	commandLogItem.Command = "COMMIT_SELL"
// 	commandLogItem.Username = user.Username
// 	commandLogItem.StockSymbol = stockSymbol
// 	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)

// 	commandLogItem.SaveRecord()
// }

// // Save a UserCommandLogItem for a COMMIT_SELL command
// func logCancelSellCommand(stockSymbol string, user *User) {
// 	commandLogItem := buildUserCommandLogItemStruct()
// 	commandLogItem.Command = "CANCEL_SELL"
// 	commandLogItem.Username = user.Username
// 	commandLogItem.StockSymbol = stockSymbol
// 	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)

// 	commandLogItem.SaveRecord()
// }
