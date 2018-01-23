package main

import (
  "log"
  "fmt"
  "net/http"
  "html/template"
  "strconv"
)

type Handler struct {}

func (handler *Handler) addHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    uID := r.Form.Get("uID")
    amount, _ := strconv.Atoi(r.Form.Get("amount"))

    u := User{Username: uID}

    var user User
    db.conn.FirstOrCreate(&user, &u)

    fmt.Fprintf(w,
      "Success!\n\n" +
      "User ID: " + uID + "\n" +
      "Amount Added: " + string(amount))
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

func (handler *Handler) quote(w http.ResponseWriter, r *http.Request) {
  temp, err := template.ParseFiles("templates/quote.html")
  if err != nil {
    log.Println("template parsing error: ", err)
  }

  err = temp.Execute(w, nil)
  if err != nil{
    log.Println("template execution error: ", err)
  }
}
