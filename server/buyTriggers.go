package main

import(
	"fmt"
	"log"
	"net/http"
)

func (handler *Handler) setBuyAmount(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		stockSymbol := r.Form.Get("stockSymbol")
		buyAmount := r.Form.Get("buyAmount")

		user, err := authUser(username)
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

		triggerQuery := BuildBuyTrigger(&user)
		triggerQuery.StockSym = stockSymbol

		var buyTrigger Trigger
		db.conn.FirstOrCreate(&buyTrigger, &triggerQuery)

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
		tx := db.conn.Begin()

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
		if err != nil {
			fmt.Fprintf(w, "Error: ", err)
			return
		}

		t := Trigger{
			UserID: user.ID,
			StockSym: stockSymbol,
		}

		var trig Trigger;
		if db.conn.First(&trig, &t).RecordNotFound() {
			fmt.Fprintf(w,
				"The trigger doesn't exist")
			return
		}

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

func (handler *Handler) setBuyTrigger(w http.ResponseWriter, r *http.Request) {
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
			Type: "buy",
		}
		var trig Trigger
		db.conn.First(&trig, &t)

		trig.PriceThreshold = thresholdInCents

		db.conn.Save(&trig)
	}
}
