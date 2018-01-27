package main

import(
	"fmt"
	"math"
	"net/http"
	"strconv"
)

func (handler *Handler) setSellAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		sellAmount := r.Form.Get("sellAmount")

		user, err := authUser(username)
		if err != nil {
			errorEventParams := map[string]string {
				"command": "SET_SELL_AMOUNT",
				"stockSymbol": stockSymbol,
				"errorMessage": err.Error(),
			}
			logErrorEvent(errorEventParams, &user)
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		// get stockHolding for this stock from user
		s := StockHolding{
			UserID: user.ID,
			StockSymbol: stockSymbol,
		}
		var stockHolding StockHolding

		// check that user has the stock holding...
		if db.conn.First(&stockHolding, &s).RecordNotFound() {
			fmt.Fprintf(w,
				"Stock holding not found...")
			return
		}

		// they want to make at least this much money
		amountToSell := stringMoneyToCents(sellAmount)

		// get the current quoted price
		quoteResponse := getQuoteFromServer(username, stockSymbol)
		quotePrice := stringMoneyToCents(quoteResponse["price"])

		numStocks, _ := convertMoneyToStock(amountToSell, quotePrice)

		if stockHolding.Number < numStocks {
			fmt.Fprintf(w,
				"Insufficient stock at the current quoted price")
			return
		}

		sellTrigger := buildSellTrigger(&user)
		sellTrigger.StockSym = stockSymbol
		sellTrigger.Amount = amountToSell
		sellTrigger.NumStocks = 0 // don't set numStocks at this point
		sellTrigger.PriceThreshold = 0
		db.conn.Create(&sellTrigger)

		fmt.Fprintf(w,
			"Sell action was successfully created!\n\n" +
			"Amount to sell: " + strconv.Itoa(int(amountToSell)) +
			"\n\nNow you should set a sell trigger...")
	}
}

func (handler *Handler) cancelSetSell(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")

		user, err := authUser(username)
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		t := Trigger{UserID: user.ID, StockSym: stockSymbol}
		var trig Trigger
		db.conn.First(&trig, &t)

		// find the stock holding associated with this
		s := StockHolding{
			UserID: user.ID,
			StockSymbol: stockSymbol,
		}
		var stockHolding StockHolding
		db.conn.First(&stockHolding, &s)

		// return the number of stocks we witheld
		stockHolding.Number += trig.NumStocks

		// ======= START TRANSACTION =======
		tx := db.conn.Begin()

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
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		threshold := r.Form.Get("threshold")

		user, err := authUser(username)
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		thresholdInCents := stringMoneyToCents(threshold);

		t := Trigger {
			UserID: user.ID,
			StockSym: stockSymbol,
			Type: "sell",
		}
		var trig Trigger
		db.conn.First(&trig, &t)

		// how much stock do we need to hit trigger amount?
		numStocks := uint(math.Ceil(float64(trig.Amount)/float64(thresholdInCents)))

		// get the stock holding
		s := StockHolding{
			UserID: user.ID,
			StockSymbol: stockSymbol,
		}
		var stockHolding StockHolding
		db.conn.First(&stockHolding, &s)

		difference := int(stockHolding.Number)-int(numStocks)
		if difference < 0 {
			// add back some stock
			stockHolding.Number += uint(int(difference)*(-1))
		} else {
			if !stockHolding.sufficient(uint(difference)) {
				return
			}
			stockHolding.Number -= uint(difference)
		}

		db.conn.Save(&stockHolding)

		// 6. Update the trigger to hold the number of stocks and set the threshold
		trig.PriceThreshold = thresholdInCents
		trig.NumStocks = numStocks
		db.conn.Save(&trig)
	}
}