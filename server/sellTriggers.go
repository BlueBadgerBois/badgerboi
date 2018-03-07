package main

import(
	"app/db"
	"fmt"
	"math"
	"net/http"
	"strconv"
)

func (handler *Handler) setSellAmount(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		sellAmount := r.Form.Get("sellAmount")

		user, err := authUser(username)
		logSetSellAmountCommand(txNum, stockSymbol, user)

		if err != nil {
			errorEventParams := map[string]string {
				"command": "SET_SELL_AMOUNT",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, &user)
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		// get stockHolding for this stock from user
		s := db.StockHolding{
			UserID: user.ID,
			StockSymbol: stockSymbol,
		}
		var stockHolding db.StockHolding

		// check that user has the stock holding...
		if dbw.Conn.First(&stockHolding, &s).RecordNotFound() {
			fmt.Fprintf(w,
				"Stock holding not found...")
			return
		}

		// they want to make at least this much money
		amountToSell := stringMoneyToCents(sellAmount)

		// get the current quoted price
		quoteResponse := getQuote(txNum, username, stockSymbol)
		quotePrice := stringMoneyToCents(quoteResponse["price"])

		numStocks, _ := convertMoneyToStock(amountToSell, quotePrice)

		if stockHolding.Number < numStocks {
			fmt.Fprintf(w,
				"Insufficient stock at the current quoted price")
			return
		}

		sellTrigger := db.BuildSellTrigger(&user)
		sellTrigger.StockSym = stockSymbol
		sellTrigger.Amount = amountToSell
		sellTrigger.NumStocks = 0 // don't set numStocks at this point
		sellTrigger.PriceThreshold = 0
		dbw.Conn.Create(&sellTrigger)

		fmt.Fprintf(w,
			"Sell action was successfully created!\n\n" +
			"Amount to sell: " + strconv.Itoa(int(amountToSell)) +
			"\n\nNow you should set a sell trigger...")
	}
}

func (handler *Handler) cancelSetSell(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")

		user, err := authUser(username)
		logCancelSetSellCommand(txNum, stockSymbol, user)

		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		trig, err := db.TriggerFromUserAndStockSym(dbw, user.ID, stockSymbol, "sell")
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			errorEventParams := map[string]string {
				"command": "CANCEL_SET_SELL",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, &user)
			return
		}

		// find the stock holding associated with this
		stockHolding, err := db.StockHoldingFromUserAndStockSym(dbw, user.ID, stockSymbol)
		if err != nil {
			fmt.Fprintf(w, "Error:", err)
			errorEventParams := map[string]string {
				"command": "CANCEL_SET_SELL",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, &user)
			return
		}

		// return the number of stocks we witheld
		stockHolding.Number += trig.NumStocks

		// ======= START TRANSACTION =======
		tx := dbw.Conn.Begin()

		if err := tx.Save(&stockHolding).Error; err != nil {
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

func (handler *Handler) setSellTrigger(w http.ResponseWriter, r *http.Request) {
	txNum := db.NewTxNum(dbw)

	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		threshold := r.Form.Get("threshold")

		user, err := authUser(username)
		logSetSellTriggerCommand(txNum, stockSymbol, user)

		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		thresholdInCents := stringMoneyToCents(threshold);

		trig, err := db.TriggerFromUserAndStockSym(dbw, user.ID, stockSymbol, "sell")
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			errorEventParams := map[string]string {
				"command": "SET_SELL_TRIGGER",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, &user)
			return
		}

		// how much stock do we need to hit trigger amount?
		numStocks := uint(math.Ceil(float64(trig.Amount)/float64(thresholdInCents)))

		// get the stock holding
		stockHolding, err := db.StockHoldingFromUserAndStockSym(dbw, user.ID, stockSymbol)
		if err != nil {
			fmt.Fprintf(w, "Error:", err)
			errorEventParams := map[string]string {
				"command": "SET_SELL_TRIGGER",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(txNum, errorEventParams, &user)
			return
		}

		difference := int(stockHolding.Number)-int(numStocks)
		if difference < 0 {
			// add back some stock
			stockHolding.Number += uint(int(difference)*(-1))
		} else {
			if !stockHolding.Sufficient(uint(difference)) {
				return
			}
			stockHolding.Number -= uint(difference)
		}

		dbw.Conn.Save(&stockHolding)

		// 6. Update the trigger to hold the number of stocks and set the threshold
		trig.PriceThreshold = thresholdInCents
		trig.NumStocks = numStocks
		dbw.Conn.Save(&trig)
	}
}

//logging functions here
func logSetSellAmountCommand(txNum uint, stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "SET_SELL_AMOUNT"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func logCancelSetSellCommand(txNum uint, stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "CANCEL_SET_SELL"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}

func logSetSellTriggerCommand(txNum uint, stockSymbol string, user db.User) {
	commandLogItem := db.BuildUserCommandLogItemStruct()
	commandLogItem.Command = "SET_SELL_TRIGGER"
	commandLogItem.Username = user.Username
	commandLogItem.StockSymbol = stockSymbol
	commandLogItem.Funds = centsToDollarsString(user.CurrentMoney)
	commandLogItem.TransactionNum = strconv.FormatUint(uint64(txNum), 10)
	username := user.Username

	commandLogItem.SaveRecord(dbw, username)
}
