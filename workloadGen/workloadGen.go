/* Workload Generator */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"net/http"
	"net/url"
	"sync"
	"time"
	"fmt"
)

const defaultTestFile = "1000User_testWorkLoad.txt"
var wg sync.WaitGroup

func serverUrl() string {
	serverUrl := "http://" + os.Getenv("WEB_URL")
	return serverUrl
}

/*
* Returns a two dimentional array, outer array contains the users,
* inner array contains the username in index [0] and his commands 
* in the remaining indexes.
*/
func divideCommandsByUser (commands []string) map[string][]string {
	commandsByUser := make(map[string][]string)
	var command []string

	for i := 0; i < len(commands); i++ {
		command = strings.Split(commands[i], ",")
		if len(command) == 0 {
			continue
		}
		user := command[1]
		//For when the command is DUMPLOG and is independent of user
		if strings.Contains(commands[i], "DUMPLOG") && len(command) == 2 {
			commandsByUser["adminDumplogs"] = append(commandsByUser["adminDumplogs"], commands[i])
			continue
		}
		commandsByUser[user] = append(commandsByUser[user], commands[i])
	}
	return commandsByUser
}

/*
*	Sends the commands in order to the server
*/
func sendUserCommands (commands []string) {
	var command []string

	for i := 0; i < len(commands); i++ {
		command = strings.Split(commands[i], ",")	
		sendCommand(command)
	}
	defer wg.Done()
}

/*
*	Sends a single command
*/
func sendCommand (command []string) {
	data := url.Values{}
	var commandType string

	if command[0] == "ADD" {
		data.Set("username", command[1])
		data.Add("amount", command[2])
		commandType = "add"
	}

	if command[0] == "QUOTE" {
		data.Set("username", command[1])
		data.Add("stocksym", command[2])
		commandType = "quote"
	}

	if command[0] == "DISPLAY_SUMMARY" {
		data.Set("username", command[1])
		commandType = "summary"	
	}

	if command[0] == "BUY" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("buyAmount", command[3])
		commandType = "buy"
	}

	if command[0] == "COMMIT_BUY" {
		data.Set("username", command[1])
		commandType = "commitBuy"
	}

	if command[0] == "CANCEL_BUY" {
		data.Set("username", command[1])
		commandType = "cancelBuy"
	}

	if command[0] == "SELL" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("sellAmount", command[3])
		commandType = "sell"
	}

	if command[0] == "COMMIT_SELL" {
		data.Set("username", command[1])
		commandType = "commitSell"
	}

	if command[0] == "CANCEL_SELL" {
		data.Set("username", command[1])
		commandType = "cancelSell"
	}

	if command[0] == "SET_BUY_AMOUNT" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("buyAmount", command[3])
		commandType = "setBuyAmount"
	}

	if command[0] == "CANCEL_SET_BUY" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		commandType = "cancelSetBuy"
	}

	if command[0] == "SET_BUY_TRIGGER" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("threshold", command[3])
		commandType = "setBuyTrigger"
	}

	if command[0] == "SET_SELL_AMOUNT" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("sellAmount", command[3])
		commandType = "setSellAmount"
	}

	if command[0] == "CANCEL_SET_SELL" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		commandType = "cancelSetSell"
	}

	if command[0] == "SET_SELL_TRIGGER" {
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("threshold", command[3])
		commandType = "setSellTrigger"
	}

	if command[0] == "DUMPLOG" {
		log.Println("IN DUMPLOG")
		var outfile string

		if(len(command) <= 2){
			data.Set("username", "admin")
			outfile = command[1]
		} else {
			data.Set("username", command[1])
			outfile = command[2]
		}
		data.Add("outfile", outfile)

		body := sendRequest(data, "dumplog")
		trimIndex := strings.Index(body, "<")
		body = body[trimIndex:]
		bodyInBytes := []byte(body)
		log.Println("writing to file now...")
		err := ioutil.WriteFile(outfile, bodyInBytes, 0644)
		if err != nil {
			log.Println("Error writing to file " + outfile)
			panic(err)
		}
		return
	}

	sendRequest(data, commandType)
}

/*
* Handles sending a request and printing the response
*/
func sendRequest (data url.Values, commandType string) string {
	var resp *http.Response
	tr := &http.Transport{
		MaxIdleConnsPerHost: 100,
		DisableKeepAlives: false,
	}

	for {
		//Set the request and reader
		fmt.Println("server url: " + serverUrl())
		req, err := http.NewRequest("POST", serverUrl() + "/" + commandType, strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Close = true

		//Set the client
		client := &http.Client{Transport: tr, Timeout: time.Duration(200 * time.Second)}
		resp, err = client.Do(req)
		if err != nil {
			log.Println("Error sending http request:")
			fmt.Println(err)
			continue
			//panic(err)
		}

		defer resp.Body.Close()
		log.Println("Response Status:", resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		return string(body)
	}
	
	return "blah"
}

func main (){
	var testFile string

	if len(os.Args) < 2 {
		log.Println("Default file used: " + defaultTestFile)
		testFile = defaultTestFile
	} else {
		testFile = os.Args[1]
	}

	fileData, err := ioutil.ReadFile(testFile)
	if err != nil {
		log.Println("Error reading file ", testFile)
		panic(err)
	}

	fileDataString := string(fileData)
	commands := strings.Split(fileDataString, "\n")

	var commandWithNum []string
	for i := 0; i < len(commands); i++ {
		if commands[i] != "" {
			commandWithNum = strings.Split(commands[i], " ")
			commands[i] = commandWithNum[1]
		}
	}

	commandsByUser := divideCommandsByUser(commands)

	
	for user,commands := range commandsByUser {
		if user != "adminDumplogs" {
			wg.Add(1)
			go sendUserCommands(commands)
		}
		// else {
			/*wg.Wait()
			wg.Add(1)
			log.Println("Admin commands")
			sendCommands(commands)
			log.Println("Almost done")
		//}*/
	}
	wg.Wait()
	wg.Add(1)
	sendUserCommands(commandsByUser["adminDumplogs"])
	log.Println("FINISHED")
}
