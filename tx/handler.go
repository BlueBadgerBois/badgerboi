package main

import (
  "log"
  "fmt"
  "net"
  "net/http"
  "html/template"
  "strconv"
  "strings"
  "bufio"
)

type Handler struct {}

func (handler *Handler) addHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    r.ParseForm()
    uID := r.Form.Get("uID")
    amount := r.Form.Get("amount")

    amountFormatted, _ := strconv.Atoi(strings.Replace(amount, ".", "", -1))

    u := User{Username: uID}

    var user User
    db.conn.FirstOrCreate(&user, &u)

    newAmount := user.Current_money + amountFormatted

    user.Current_money = newAmount
    db.conn.Save(&user)

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

  if r.Method == "POST" {
    r.ParseForm()
    uID := r.Form.Get("uID")
    stocksym := r.Form.Get("stocksym")

    log.Println("UID: " + uID + ", Stock: " + stocksym)

    //hit quote server
    conn, err := net.Dial("tcp", "192.168.0.29:3333")

    if err != nil{
    	log.Println("error hitting quote server: ", err)
    }

    fmt.Fprintf(conn, uID + ", " + stocksym)
    message, _ := bufio.NewReader(conn).ReadString('\n')

    fmt.Fprintf(w,
      "Success!\n\n" +
      "Quote Server Response: " + message)
  }
}
