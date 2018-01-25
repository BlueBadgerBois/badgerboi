package main

import (
	"github.com/jinzhu/gorm"
)

type LogItem struct {
	gorm.Model
	data string // JSON containing the other attributes
}
