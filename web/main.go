package main

import (
	"html/template"
	"github.com/kabukky/httpscerts" //might download later for faster import OR implement own version of cert generation
	"net/http"
	"log"
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

func main() {
	// Check if the cert files are available.
  err := httpscerts.Check("cert.pem", "key.pem")
  // If they are not available, generate new ones.
  if err != nil {
    err = httpscerts.Generate("cert.pem", "key.pem", "127.0.0.1:8081")
    if err != nil {
      log.Fatal("Error: Couldn't create https certs.")
    }
  }

	http.HandleFunc("/", handler)
	http.ListenAndServeTLS(":8080", "cert.pem", "key.pem", nil)
	
	//redirect http to https here
}