package main

import (
	"errors"
	"fmt"
	"net/http"
)

func (handler *Handler) buy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		buyAmount := r.Form.Get("buyAmount")
		stockSymbol := r.Form.Get("stockSymbol")

		amountToBuyInCents := stringMoneyToCents(buyAmount)

		user := UserFromUsernameOrCreate(db, username)

		logBuyCommand(stockSymbol, user)

		quoteResponseMap := getQuoteFromServer(username, stockSymbol)

		quotedPriceInCents := stringMoneyToCents(quoteResponseMap["price"])

		if !anyStocksCanBeBought(amountToBuyInCents, quotedPriceInCents) {
			errMsg := "Error! The buy amount " + centsToDollarsString(amountToBuyInCents) +
			" is not enough to buy even one stock at price " + centsToDollarsString(quotedPriceInCents)
			fmt.Fprintf(w, errMsg)
			errorEventParams := map[string]string {
				"command": "BUY",
				"stockSymbol": stockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(errorEventParams, user)
			return
		}

		err, _ := createBuyTransaction(user, stockSymbol, amountToBuyInCents, quoteResponseMap["price"])

		if err != nil {
			fmt.Fprintf(w, "Error!" + err.Error())
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

func (handler *Handler) commitBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")

		user, err := UserFromUsername(db, username)
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

		transactionToCommit, err := NewestPendingTransactionForUser(db, user, "buy")
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

		logCommitBuyCommand(transactionToCommit.StockSymbol, user)


		// later we need to make the job server just update the quote price
		currentTime := db.getCurrentTime()
		timeDifference := currentTime.Sub(transactionToCommit.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			fmt.Fprintf(w, "Failure! Most recent buy transaction is more than 60 seconds old.")
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": "Failure! Buy transaction is more than 60 seconds old.",
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		tx := db.Begin()

		// UserFromUsername(tx, user.Username) // reload user
		tx.conn.Where(&User{Username: user.Username}).First(&user) // reload user

		numStocksToBuy, leftOverCents := convertMoneyToStock(transactionToCommit.AmountInCents,
		transactionToCommit.QuotedStockPrice)

		moneyToWithDraw := transactionToCommit.AmountInCents - leftOverCents

		if !user.HasEnoughMoney(moneyToWithDraw) {
			tx.Rollback() // rollback immediately

			errMsg := "Failure! User does not have enough money\n\n" +
			"User ID: " + user.Username + "\n" +
			"Stock Symbol: " + transactionToCommit.StockSymbol + "\n" +
			"Number of stocks to buy: " + string(numStocksToBuy) +
			"Cost to user: " + centsToDollarsString(moneyToWithDraw) +
			"\nCurrent funds: " + centsToDollarsString(user.CurrentMoney)

			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": "Insufficient funds",
			}
			fmt.Fprintf(w, errMsg)
			logErrorEvent(errorEventParams, user)
			return
		}

		// give user stocks
		var stockHolding StockHolding
		holdingQuery := StockHolding{
			UserID: user.ID,
			StockSymbol: transactionToCommit.StockSymbol,
		}
		tx.conn.FirstOrCreate(&stockHolding, &holdingQuery)

		stockHolding.Number += numStocksToBuy
		tx.conn.Save(&stockHolding)

		// Subtract money from user's account
		user.CurrentMoney -= moneyToWithDraw
		tx.conn.Save(&user)

		transactionToCommit.State = "complete"
		tx.conn.Save(&transactionToCommit)

		tx.conn.Commit()

		fmt.Fprintf(w, "BUY committed.\n " +
		"symbol: " + transactionToCommit.StockSymbol +
		"\nnum bought: " + string(numStocksToBuy) +
		"\nquoted price: " + centsToDollarsString(transactionToCommit.QuotedStockPrice))
	}
}

func (handler *Handler) cancelBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")


		user, err := UserFromUsername(db, username)
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

		transactionToCancel, err := NewestPendingTransactionForUser(db, user, "buy")
		if err != nil {
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "CANCEL_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		logCancelBuyCommand(transactionToCancel.StockSymbol, user)

		// later we need to make the job server just update the quote price
		currentTime := db.getCurrentTime()
		timeDifference := currentTime.Sub(transactionToCancel.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			errMsg := "Failure! Most recent BUY transaction is more than 60 seconds old."
			fmt.Fprintf(w, errMsg)
			errorEventParams := map[string]string {
				"command": "CANCEL_BUY",
				"stockSymbol": transactionToCancel.StockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(errorEventParams, user)

			return
		}

		transactionToCancel.Cancel(db)

		successMsg := "BUY cancelled.\n " +
		"symbol: " + transactionToCancel.StockSymbol +
		"\nmoney that was to be committed: " + centsToDollarsString(transactionToCancel.AmountInCents)
		fmt.Fprintf(w, successMsg)
	}
}

func createBuyTransaction(user *User, stockSymbol string, amountToBuyInCents uint, quotePrice string) (error, *Transaction) {
	// create a transaction record
	buyTransaction := BuildBuyTransaction(user)
	buyTransaction.StockSymbol = stockSymbol
	buyTransaction.AmountInCents = amountToBuyInCents
	buyTransaction.QuotedStockPrice = stringMoneyToCents(quotePrice)
	db.conn.NewRecord(buyTransaction)

	if !user.HasEnoughMoney(amountToBuyInCents) {
		errorEventParams := map[string]string {
			"command": "BUY",
			"stockSymbol": stockSymbol,
			"errorMessage": "Insufficient funds",
		}
		logErrorEvent(errorEventParams, user)

		return errors.New("Failure! User does not have enough money\n\n" +
		"User ID: " + user.Username + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Amount to buy: " + centsToDollarsString(amountToBuyInCents) +
		"\nCurrent funds: " + centsToDollarsString(user.CurrentMoney)),
		buyTransaction
	}
	db.conn.Save(buyTransaction)
	return nil, buyTransaction
}

// Save a UserCommandLogItem for a BUY command
func logBuyCommand(stockSymbol string, user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

// Save a UserCommandLogItem for a COMMIT_BUY command
func logCommitBuyCommand(stockSymbol string, user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "COMMIT_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username

	commandLogItem.SaveRecord(username)
}

// Save a UserCommandLogItem for a COMMIT_BUY command
func logCancelBuyCommand(stockSymbol string, user *User) {
	commandLogItem := BuildUserCommandLogItemStruct()
	commandLogItem.Command = "CANCEL_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	username := user.Username


	commandLogItem.SaveRecord(username)
}
