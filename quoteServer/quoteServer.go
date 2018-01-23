package main

import (
    "fmt"
    "net"
    "os"
    "strings"
    "math/rand"
    "time"
    "bytes"
    "strconv"
)

const (
    CONN_HOST = "localhost"
    CONN_PORT = "3333"
    CONN_TYPE = "tcp"
)

type StockResponse struct {
  Quote     float64 `json:"quote"`
  Symbol    string  `json:"symbol"`
  UserID    string  `json:"userID"`
  Timestamp string  `json:"timestamp"`
  Cryptokey string  `json:"cryptokey"`
}

func getStockResponse(buf []byte) string {
  s := string(bytes.Trim(buf, "\x00"))
  sp := strings.Split(s, ",")
  // for now assume we will never get malformed inputs and this limited error check and crashing is ok
  if(len(sp) != 2) {
    fmt.Print("Input not in the expected form\n")
    os.Exit(1)
  }

  var quoteRes string = strconv.Itoa(rand.Intn(999)) + "." + strconv.Itoa(rand.Intn(99)) +	//Quote
  	", " + sp[0] +	//Symbol
  	", " + sp[1] +	//UserID
  	", " + time.Now().String() + //Timestamp
  	", abc"	//CryptoKey

  return quoteRes
}

// Handles incoming requests.
func handleRequest(conn net.Conn) {
  buf := make([]byte, 1024)
  _, err := conn.Read(buf)
  if err != nil {
    fmt.Println("Error reading:", err.Error())
  }

  stockResponse := getStockResponse(buf)


  // Send a response back to person contacting us.
  conn.Write([]byte(stockResponse))
  conn.Close()
}

func main() {
    // Listen for incoming connections.
    listener, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)
    if err != nil {
        fmt.Println("Error listening:", err.Error())
        os.Exit(1)
    }
    defer listener.Close()
    fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

    for {
        conn, err := listener.Accept()
        if err != nil {
            fmt.Println("Error accepting: ", err.Error())
            os.Exit(1)
        }
        go handleRequest(conn)
    }
}
