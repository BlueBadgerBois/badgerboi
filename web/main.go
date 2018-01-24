package main

import (
  "html/template"
  "github.com/kabukky/httpscerts" //might download later for faster import OR implement own version of cert generation
  "net/http"
  "log"
  "fmt"
)

func handler(w http.ResponseWriter, r *http.Request){
  temp, err := template.ParseFiles("templates/login.html")
  if err != nil {
    log.Println("template parsing error: ", err)
  }

  p := "hey" //filler data so function Execute executes
  err = temp.Execute(w, p)
  if err != nil{
    log.Println("template execution error: ", err)
  }

  log.Println("Login Page rendered.")
}

func addHandler(w http.ResponseWriter, r *http.Request) {
  if r.Method == "POST" {
    type addJson struct {
      UID string
      Amount string
    }
    r.ParseForm()
    addReq := addJson{
      UID: r.Form.Get("uID"),
      Amount: r.Form.Get("amount"),
    }
    fmt.Fprintf(w,
      "User ID: " + addReq.UID + "\n" +
      "Amount Added: " + addReq.Amount)
  }
}

func main() {
  // Check if the cert files are available.
  err := httpscerts.Check("cert.pem", "key.pem")
  // If they are not available, generate new ones.
  if err != nil {
    err = httpscerts.Generate("cert.pem", "key.pem", "web:8081")
    if err != nil {
      log.Fatal("Error: Couldn't create https certs.")
    }
  }

  http.HandleFunc("/", handler)
  http.HandleFunc("/add", addHandler)
  http.ListenAndServeTLS(":8081", "cert.pem", "key.pem", nil)
}
