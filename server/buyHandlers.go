package main

import (
	"app/db"
	"errors"
	"fmt"
	"net/http"
)

func (handler *Handler) buy(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		buyAmount := r.Form.Get("buyAmount")
		stockSymbol := r.Form.Get("stockSymbol")

		amountToBuyInCents := stringMoneyToCents(buyAmount)

		user := db.UserFromUsernameOrCreate(dbw, username)

		logBuyCommand(txNum, stockSymbol, user)

		quoteResponseMap := getQuoteFromServer(txNum, username, stockSymbol)

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
			logErrorEvent(txNum, errorEventParams, user)
			return
		}

		err, _ := createBuyTransaction(txNum, user, stockSymbol, amountToBuyInCents, quoteResponseMap["price"])

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
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")

		user, err := db.UserFromUsername(dbw, username)
		if err != nil {
			fmt.Fprintf(w, "Failure! user does not exist!\n\n")
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, user)

			return
		}

		transactionToCommit, err := db.NewestPendingTransactionForUser(dbw, user, "buy")
		if err != nil {
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, user)
			return
		}

		logCommitBuyCommand(txNum, transactionToCommit.StockSymbol, user)


		// later we need to make the job server just update the quote price
		currentTime := dbw.GetCurrentTime()
		timeDifference := currentTime.Sub(transactionToCommit.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			fmt.Fprintf(w, "Failure! Most recent buy transaction is more than 60 seconds old.")
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": transactionToCommit.StockSymbol,
				"errorMessage": "Failure! Buy transaction is more than 60 seconds old.",
			}
			logErrorEvent(txNum, errorEventParams, user)

			return
		}

		tx := dbw.Begin()

		// UserFromUsername(tx, user.Username) // reload user
		tx.Conn.Where(&db.User{Username: user.Username}).First(&user) // reload user

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
			logErrorEvent(txNum, errorEventParams, user)
			return
		}

		// give user stocks
		var stockHolding db.StockHolding
		holdingQuery := db.StockHolding{
			UserID: user.ID,
			StockSymbol: transactionToCommit.StockSymbol,
		}
		tx.Conn.FirstOrCreate(&stockHolding, &holdingQuery)

		stockHolding.Number += numStocksToBuy
		tx.Conn.Save(&stockHolding)

		// Subtract money from user's account
		user.CurrentMoney -= moneyToWithDraw
		tx.Conn.Save(&user)

		transactionToCommit.State = "complete"
		tx.Conn.Save(&transactionToCommit)

		tx.Conn.Commit()

		fmt.Fprintf(w, "BUY committed.\n" +
		"symbol: " + transactionToCommit.StockSymbol +
		"\nnum bought: " + fmt.Sprintf("%v", numStocksToBuy) +
		"\nquoted price: " + centsToDollarsString(transactionToCommit.QuotedStockPrice))
	}
}

func (handler *Handler) cancelBuy(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")


		user, err := db.UserFromUsername(dbw, username)
		if err != nil {
			fmt.Fprintf(w, "Failure! user does not exist!\n\n")
			errorEventParams := map[string]string {
				"command": "COMMIT_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, user)

			return
		}

		transactionToCancel, err := db.NewestPendingTransactionForUser(dbw, user, "buy")
		if err != nil {
			fmt.Fprintf(w, "Failure! " + err.Error())
			errorEventParams := map[string]string {
				"command": "CANCEL_BUY",
				"stockSymbol": "",
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, user)

			return
		}

		logCancelBuyCommand(txNum, transactionToCancel.StockSymbol, user)

		// later we need to make the job server just update the quote price
		currentTime := dbw.GetCurrentTime()
		timeDifference := currentTime.Sub(transactionToCancel.CreatedAt)
		if timeDifference.Seconds() > MAX_TRANSACTION_VALIDITY_SECS {
			errMsg := "Failure! Most recent BUY transaction is more than 60 seconds old."
			fmt.Fprintf(w, errMsg)
			errorEventParams := map[string]string {
				"command": "CANCEL_BUY",
				"stockSymbol": transactionToCancel.StockSymbol,
				"errorMessage": errMsg,
			}
			logErrorEvent(txNum, errorEventParams, user)

			return
		}

		transactionToCancel.Cancel(dbw)

		successMsg := "BUY cancelled.\n " +
		"symbol: " + transactionToCancel.StockSymbol +
		"\nmoney that was to be committed: " + centsToDollarsString(transactionToCancel.AmountInCents)
		fmt.Fprintf(w, successMsg)
	}
}

func createBuyTransaction(txNum uint, user *db.User, stockSymbol string, amountToBuyInCents uint, quotePrice string) (error, *db.Transaction) {

	// create a transaction record
	buyTransaction := db.BuildBuyTransaction(user)
	buyTransaction.StockSymbol = stockSymbol
	buyTransaction.AmountInCents = amountToBuyInCents
	buyTransaction.QuotedStockPrice = stringMoneyToCents(quotePrice)
	dbw.Conn.NewRecord(buyTransaction)

	if !user.HasEnoughMoney(amountToBuyInCents) {
		errorEventParams := map[string]string {
			"command": "BUY",
			"stockSymbol": stockSymbol,
			"errorMessage": "Insufficient funds",
		}
		logErrorEvent(txNum, errorEventParams, user)

		return errors.New("Failure! User does not have enough money\n\n" +
		"User ID: " + user.Username + "\n" +
		"Stock Symbol: " + stockSymbol + "\n" +
		"Amount to buy: " + centsToDollarsString(amountToBuyInCents) +
		"\nCurrent funds: " + centsToDollarsString(user.CurrentMoney)),
		buyTransaction
	}
	dbw.Conn.Save(buyTransaction)
	return nil, buyTransaction
}

// Save a UserCommandLogItem for a BUY command
func logBuyCommand(txNum uint, stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = string(txNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for a COMMIT_BUY command
func logCommitBuyCommand(txNum uint, stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "COMMIT_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = string(txNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

// Save a UserCommandLogItem for a COMMIT_BUY command
func logCancelBuyCommand(txNum uint, stockSymbol string, user *db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "CANCEL_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = string(txNum)
	username := user.Username


	commandLogItem.SaveRecord(dbw, username)
}
