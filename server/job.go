package main

import (
	"log"
	"time"
)

func runJobServer() {
	tickChan := time.NewTicker(5*time.Second).C
	for {
		<- tickChan
		log.Println("Checking all triggers")
		checkTriggers()
	}
}

func checkTriggers() {
	var triggers []Trigger
	db.conn.Find(&triggers)
	for _, trig := range triggers {
		var user User
		db.conn.First(&user)

		responseMap := getQuoteFromServer(user.Username, trig.StockSym)

		quotePrice := stringMoneyToCents(responseMap["price"])
		if trig.Type == "buy" && quotePrice < trig.PriceThreshold {
			amntToBuy, leftover := convertMoneyToStock(trig.Amount, quotePrice)

			stockHolding := StockHolding{
				UserID: user.ID,
				StockSymbol: trig.StockSym,
				Number: amntToBuy,
			}
			user.CurrentMoney = user.CurrentMoney + leftover

			// ======== START TRANSACTION ========
			tx := db.conn.Begin()

			if err := tx.Create(&stockHolding).Error; err != nil {
				tx.Rollback()
				continue
			}

			if err := tx.Save(&user).Error; err != nil {
				tx.Rollback()
				continue
			}


			if err := tx.Delete(&trig).Error; err != nil {
				tx.Rollback()
				continue
			}

			tx.Commit()
			// ========= END TRANSACTION =========

		} else if trig.Type == "sell" && quotePrice > trig.PriceThreshold {
			log.Println("We are above the threshold so we can sell!")
		} else {
			log.Println("trigger not executing")
		}
	}
}