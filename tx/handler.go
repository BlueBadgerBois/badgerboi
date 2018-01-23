package main

import (
  "log"
  "net/http"
  "html/template"
)

type Handler struct {}

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
