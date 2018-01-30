package main

import (
	"encoding/json"
	"log"
	"github.com/jinzhu/gorm"
)

var USER_COMMAND_TYPES = [...]string{
	"ADD",
	"QUOTE",
	"BUY",
	"COMMIT_BUY",
	"CANCEL_BUY",
	"SELL",
	"COMMIT_SELL",
	"CANCEL_SELL",
	"SET_BUY_AMOUNT",
	"CANCEL_SET_BUY",
	"SET_BUY_TRIGGER",
	"SET_SELL_AMOUNT",
	"SET_SELL_TRIGGER",
	"CANCEL_SET_SELL",
	"DUMPLOG",
	"DISPLAY_SUMMARY",
}

// TransactionNum will be derived from the id column (which is autoincremented)
// Timestamp will be derived from the created_at column
// These will be derived when dumping the logs
type LogItem struct {
	gorm.Model
	Username string
	Data string // JSON containing the other attributes
}


// The structs below are to structure the data string for a LogItem.
// Marshal to JSON string before passing to a new LogItem.

type UserCommandLogItem struct {
	// User commands come from the user command files or from manual entries in
	// the students' web forms
	LogType string // 'UserCommandType'
	Server string
	Command string
	Username string
	StockSymbol string // what to do for summary?
	Filename string
	Funds string // should be in dollars
}

func BuildUserCommandLogItemStruct() UserCommandLogItem {
	logItem := UserCommandLogItem {
		LogType: "UserCommandType",
		Server: "someServer",
		Filename: "inputFile", // this needs to be changed later to be overridden with an actual file name
	}
		return logItem
}

func (logItem *UserCommandLogItem) SaveRecord(username string) {
	jsonBytes, err := json.Marshal(logItem)

	if err != nil { log.Fatal("Unable to convert struct %s to json", logItem) }

	jsonString := bytesToString(jsonBytes)

	saveLogItem(db, username, jsonString)
}

type QuoteServerLogItem struct {
	// Every hit to the quote server requires a log entry with the results. The
	// price, symbol, username, timestamp and cryptokey are as returned by the quote server
	LogType string // 'QuoteServerType'
	Server string
	Price string
	StockSymbol string
	Username string
	QuoteServerTime string
	Cryptokey string
}

func BuildQuoteServerLogItemStruct() QuoteServerLogItem {
	logItem := QuoteServerLogItem {
		LogType: "QuoteServerType",
		Server: "someServer",
	}
	return logItem
}

func (logItem *QuoteServerLogItem) SaveRecord(username string) {
	jsonBytes, err := json.Marshal(logItem)

	if err != nil { log.Fatal("Unable to convert struct %s to json", logItem) }

	jsonString := bytesToString(jsonBytes)

	saveLogItem(db, username, jsonString)
}


type AccountTransactionLogItem struct {
	// Any time a user's account is touched, an account message is printed.
	// Appropriate actions are "add" or "remove".
	LogType string // 'AccountTransactionType'
	Server string
	Action string
	Username string
	Funds string // dollars
}

func BuildAccountTransactionLogItemStruct() AccountTransactionLogItem {
	logItem := AccountTransactionLogItem {
		LogType: "AcountTransactionType",
		Server: "someServer",
	}
	return logItem
}

func (logItem *AccountTransactionLogItem) SaveRecord(username string) {
	jsonBytes, err := json.Marshal(logItem)

	if err != nil { log.Fatal("Unable to convert struct %s to json", logItem) }

	jsonString := bytesToString(jsonBytes)

	saveLogItem(db, username, jsonString)
}

type SystemEventLogItem struct {
	// System events can be current user commands, interserver communications,
	// or the execution of previously set triggers
	LogType string // 'SystemEventType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds string // dollars
}

func BuildSystemEventLogItemStruct() SystemEventLogItem {
	logItem := SystemEventLogItem {
		LogType: "SystemEventType",
		Server: "someServer",
	}
	return logItem
}

func (logItem *SystemEventLogItem) SaveRecord(username string) {
	jsonBytes, err := json.Marshal(logItem)

	if err != nil { log.Fatal("Unable to convert struct %s to json", logItem) }

	jsonString := bytesToString(jsonBytes)

	saveLogItem(db, username, jsonString)
}

type ErrorEventLogItem struct {
	// Error messages contain all the information of user commands, in
	// addition to an optional error message
	LogType string // 'ErrorEventType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds string // dollars
	ErrorMessage string
}

func BuildErrorEventLogItemStruct() ErrorEventLogItem {
	logItem := ErrorEventLogItem {
		LogType: "ErrorEventType",
		Server: "someServer",
		Filename: "inputFile", // this needs to be changed later to be overridden with an actual file name
	}
	return logItem
}

func (logItem *ErrorEventLogItem) SaveRecord(username string) {
	jsonBytes, err := json.Marshal(logItem)

	if err != nil { log.Fatal("Unable to convert struct %s to json", logItem) }

	jsonString := bytesToString(jsonBytes)

	saveLogItem(db, username, jsonString)
}

type DebugLogItem struct {
	// Debugging messages contain all the information of user commands, in
	// addition to an optional debug message
	LogType string // 'DebugType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds string // dollars
	DebugMessage string
}

func bytesToString(bytes []byte) string {
	return string(bytes[:])
}

func saveLogItem(db *DBW, username string, jsonData string) {
	logItem := LogItem{Username: username, Data: jsonData}
	db.conn.NewRecord(logItem)
	db.conn.Create(&logItem)
}
