package main

import (
	"app/db"
	"fmt"
	"log"
	"time"
)

func runJobServer() {
	tickChan := time.NewTicker(60*time.Second).C
	for {
		<- tickChan
		log.Println("Checking all triggers")
		checkTriggers()
	}
}

func checkTriggers() {
	var triggers []db.Trigger
	dbw.Conn.Find(&triggers)
	for _, trig := range triggers {
		var user db.User
		dbw.Conn.Where("ID = ?", trig.UserID).First(&user)

		txNum, err := db.FirstTxNum(dbw)
		if err != nil {
			fmt.Println("Error: ", err.Error())
			return
		}
		responseMap := getQuoteFromServer(txNum, user.Username, trig.StockSym)

		quotePrice := stringMoneyToCents(responseMap["price"])
		if trig.Type == "buy" && quotePrice < trig.PriceThreshold {

			amntToBuy, leftover := convertMoneyToStock(trig.Amount, quotePrice)

			stockHolding := db.StockHolding{
				UserID: user.ID,
				StockSymbol: trig.StockSym,
			}
			user.CurrentMoney = user.CurrentMoney + leftover

			// ======== START TRANSACTION ========
			tx := dbw.Conn.Begin()

			if err := tx.FirstOrCreate(&stockHolding).Error; err != nil {
				tx.Rollback()
				continue
			}

			stockHolding.Number += amntToBuy
			tx.Save(stockHolding)

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

		} else if trig.Type == "sell" &&
				quotePrice > trig.PriceThreshold &&
				trig.PriceThreshold != 0 {

			// we need to update user's money
			user.CurrentMoney += (trig.NumStocks * quotePrice)

			// ======== START TRANSACTION ========
			tx := dbw.Conn.Begin()

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


		} else {
			log.Println("trigger not executing")
		}
	}
}
