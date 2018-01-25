package main

import (
	"github.com/jinzhu/gorm"
)

var COMMAND_TYPES = [...]string{
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
	data string // JSON containing the other attributes
}

// The structs below are tostructure the data string for a LogItem.
// Marshal to JSON string before passing to a new LogItem.

// User commands come from the user command files or from manual entries in
// the students' web forms
type UserCommandLogItem struct {
	LogType string // 'UserCommandType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds uint
}

// Every hit to the quote server requires a log entry with the results. The
// price, symbol, username, timestamp and cryptokey are as returned by the quote server
type QuoteLogItem struct {
	LogType string // 'QuoteServerType'
	Server string
	Price string
	StockSymbol string
	Username string
	QuoteServerTime string
	Cryptokey string
}

// Any time a user's account is touched, an account message is printed.
// Appropriate actions are "add" or "remove".
type AccountTransactionLogItem struct {
	LogType string // 'AccountTransactionType'
	Server string
	Action string
	Username string
	Funds uint
}

// System events can be current user commands, interserver communications,
// or the execution of previously set triggers
type SystemEventLogItem struct {
	LogType string // 'SystemEventType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds uint
}

// Error messages contain all the information of user commands, in
// addition to an optional error message
type ErrorEventLogItem struct {
	LogType string // 'ErrorEventType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds uint
	ErrorMessage string
}

// Debugging messages contain all the information of user commands, in
// addition to an optional debug message
type DebugLogItem struct {
	LogType string // 'DebugType'
	Server string
	Command string
	Username string
	StockSymbol string
	Filename string
	Funds uint
	DebugMessage string
}

