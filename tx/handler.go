package main

import (
  "log"
  "fmt"
  "net/http"
  "html/template"
)

type Handler struct {}

func (handler *Handler) addHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    uID := r.Form.Get("uID")
    amount := r.Form.Get("amount")

    log.Println("UID: " + uID + ", Amount: " + amount)
    fmt.Fprintf(w,
      "Success!\n\n" +
      "User ID: " + uID + "\n" +
      "Amount Added: " + amount)
  }
}

func (handler *Handler) index(w http.ResponseWriter, r *http.Request) {
  temp, err := template.ParseFiles("templates/login.html")
  if err != nil {
    log.Println("template parsing error: ", err)
  }

  err = temp.Execute(w, nil)
  if err != nil{
    log.Println("template execution error: ", err)
  }
}
