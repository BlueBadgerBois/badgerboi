package main

import(
	"app/db"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (handler *Handler) setBuyAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		buyAmount := r.Form.Get("buyAmount")

		user, err := authUser(username)
		logSetBuyAmountCommand(stockSymbol, user)

		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			errorEventParams := map[string]string {
				"command": "SET_BUY_AMOUNT",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, &user)
			return
		}

		amountToBuyInCents := stringMoneyToCents(buyAmount)

		triggerQuery := db.BuildBuyTrigger(&user)
		triggerQuery.StockSym = stockSymbol

		var buyTrigger db.Trigger
		dbw.Conn.FirstOrCreate(&buyTrigger, &triggerQuery)

		oldAmount := buyTrigger.Amount
		newAmount := amountToBuyInCents

		amountDifference := int(newAmount)-int(oldAmount)
		log.Println(amountDifference)

		if amountDifference < 0 {
			user.CurrentMoney += uint(amountDifference*(-1))
		} else if !user.HasEnoughMoney(newAmount-oldAmount) {
			fmt.Fprintf(w, "Insufficient funds")
			errorEventParams := map[string]string {
				"command": "SET_BUY_AMOUNT",
				"stockSymbol": stockSymbol,
				"buyAmount" : buyAmount,
				"errorMessage": "Insufficient funds",
			}
			logErrorEvent(errorEventParams, &user)
			return
		} else {
			user.CurrentMoney -= uint(amountDifference)
		}

		buyTrigger.Amount = uint(amountToBuyInCents)

		// ====== START TRANSACTION ======
		tx := dbw.Conn.Begin()

		if err := tx.Save(&user).Error; err != nil {
			tx.Rollback()
			return
		}

		if err := tx.Save(&buyTrigger).Error; err != nil {
			tx.Rollback()
			return
		}

		tx.Commit()
		// ======= END TRANSACTION =======

		fmt.Fprintf(w,
			"Buy action was successfully created!\n\n" +
			"Amount withdrawn: $" + centsToDollarsString(amountToBuyInCents) +
			"\n\nNow you should set a trigger...")
	}
}

func (handler *Handler) cancelSetBuy(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")

		user, err := authUser(username)
		logCancelSetBuyCommand(stockSymbol, user)

		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		trig, err := db.TriggerFromUserAndStockSym(dbw, user.ID, stockSymbol, "buy")
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			errorEventParams := map[string]string {
				"command": "BUY",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, &user)
			return
		}

		user.CurrentMoney = user.CurrentMoney + trig.Amount

		// ======= START TRANSACTION =======
		tx := dbw.Conn.Begin()

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

func (handler *Handler) setBuyTrigger(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		threshold := r.Form.Get("threshold")

		user, err := authUser(username)
		logSetBuyTriggerCommand(stockSymbol, user)

		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		thresholdInCents := stringMoneyToCents(threshold);

		trig, err := db.TriggerFromUserAndStockSym(dbw, user.ID, stockSymbol, "buy")
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			errorEventParams := map[string]string {
				"command": "BUY",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, &user)
			return
		}

		trig.PriceThreshold = thresholdInCents

		dbw.Conn.Save(&trig)
	}
}

//logging functions here
func logSetBuyAmountCommand(stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "SET_BUY_AMOUNT"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func logCancelSetBuyCommand(stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "CANCEL_SET_BUY"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func logSetBuyTriggerCommand(stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "SET_BUY_TRIGGER"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.Itoa(currentTxNum)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}
