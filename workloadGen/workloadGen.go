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
)

const serverUrl = "http://web:8082"
const defaultTestFile = "10User_testWorkLoad.txt"
var wg sync.WaitGroup

/*
* Returns a two dimentional array, outer array contains the users,
* inner array contains the username in index [0] and his commands 
* in the remaining indexes.
*/
func divideCommandsByUser (commands []string) [][]string {
	log.Println("In divideCommandsByUser")
	commandsByUser := make([][]string, 0)
	var command []string
	foundDumplog := false
	foundUsername := false

	for i := 0; i < len(commands); i++ {
		command = strings.Split(commands[i], ",")

		//For when the command is DUMPLOG and is independent of user
		if strings.Contains(commands[i], "DUMPLOG") && len(command) == 2 {
			for j := 0; j < len(commandsByUser); j++ {
				if commandsByUser[j][0] == "adminDumplogs" {
					commandsByUser[j] = append(commandsByUser[j], commands[i])
					foundDumplog = true
				}
			}

			if !foundDumplog {
				commandsByUser = append(commandsByUser, make([]string, 0))
				commandsByUser[len(commandsByUser) - 1] = append(commandsByUser[len(commandsByUser) - 1], "adminDumplogs")
				commandsByUser[len(commandsByUser) - 1] = append(commandsByUser[len(commandsByUser) - 1], commands[i])
			}
			foundDumplog = false
			continue
		}

		//Rest of commands that depend on user
		for j := 0; j < len(commandsByUser); j++ {
			if commandsByUser[j][0] == command[1] {
				commandsByUser[j] = append(commandsByUser[j], commands[i])
				foundUsername = true
			}
		}
		if !foundUsername {
			commandsByUser = append(commandsByUser, make([]string, 0))
			commandsByUser[len(commandsByUser) - 1] = append(commandsByUser[len(commandsByUser) - 1], command[1])
			commandsByUser[len(commandsByUser) - 1] = append(commandsByUser[len(commandsByUser) - 1], commands[i])
		}
		foundUsername = false
	}

	return commandsByUser
}

/*
*	Sends the commands in order to the server
*/
func sendCommands (commands []string) {
	log.Println("IN Send Commands")
	var command []string

	for i := 1; i < len(commands); i++ {
		command = strings.Split(commands[i], ",")
		log.Println("----------------------------------------------")
		log.Println(command)		
		sendCommand(command)
	}
	defer wg.Done()
}

/*
*	Sends a single command
*/
func sendCommand (command []string) {
	if command[0] == "ADD" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("amount", command[2])
		req, err := http.NewRequest("POST", serverUrl + "/add", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "QUOTE" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stocksym", command[2])
		req, err := http.NewRequest("POST", serverUrl + "/quote", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "DISPLAY_SUMMARY" {
		data := url.Values{}
		data.Set("username", command[1])
		req, err := http.NewRequest("POST", serverUrl + "/summary", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "BUY" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("buyAmount", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/buy", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "COMMIT_BUY" {
		data := url.Values{}
		data.Set("username", command[1])
		req, err := http.NewRequest("POST", serverUrl + "/commitBuy", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "CANCEL_BUY" {
		data := url.Values{}
		data.Set("username", command[1])
		req, err := http.NewRequest("POST", serverUrl + "/cancelBuy", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "SELL" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("sellAmount", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/sell", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "COMMIT_SELL" {
		data := url.Values{}
		data.Set("username", command[1])
		req, err := http.NewRequest("POST", serverUrl + "/commitSell", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "CANCEL_SELL" {
		data := url.Values{}
		data.Set("username", command[1])
		req, err := http.NewRequest("POST", serverUrl + "/cancelSell", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "SET_BUY_AMOUNT" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("buyAmount", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/setBuyAmount", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "CANCEL_SET_BUY" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		req, err := http.NewRequest("POST", serverUrl + "/cancelSetBuy", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "SET_BUY_TRIGGER" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("threshold", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/setBuyTrigger", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "SET_SELL_AMOUNT" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("sellAmount", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/setSellAmount", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "CANCEL_SET_SELL" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		req, err := http.NewRequest("POST", serverUrl + "/cancelSetSell", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "SET_SELL_TRIGGER" {
		data := url.Values{}
		data.Set("username", command[1])
		data.Add("stockSymbol", command[2])
		data.Add("threshold", command[3])
		req, err := http.NewRequest("POST", serverUrl + "/setSellTrigger", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}
		sendRequest(req)
	}

	if command[0] == "DUMPLOG" {
		log.Println("In dumplog")
		data := url.Values{}
		var outfile string

		if(len(command) <= 2){
			data.Set("username", "admin")
			outfile = command[1]
		} else {
			data.Set("username", command[1])
			outfile = command[2]
		}
		data.Add("outfile", outfile)

		req, err := http.NewRequest("POST", serverUrl + "/dumplog", strings.NewReader(data.Encode()))
		if err != nil {
			log.Println("Error making a new request:")
			log.Println(err)
		}

		body := sendRequest(req)
		trimIndex := strings.Index(body, "<")
		body = body[trimIndex:]
		bodyInBytes := []byte(body)
		err = ioutil.WriteFile(outfile, bodyInBytes, 0644)
		if err != nil {
			log.Println("Error writing to file " + outfile)
			panic(err)
		}
	}
}

/*
* Handles sending a request and printing the response
*/
func sendRequest (req *http.Request) string {
	log.Println("In send request")
	log.Println("Sending request...")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending http request:")
		panic(err)
	}
	defer resp.Body.Close()
	log.Println("Response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("Response Body:", string(body))
	return string(body)
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
		commandWithNum = strings.Split(commands[i], " ")
		commands[i] = commandWithNum[1]
	}

	commandsByUser := divideCommandsByUser(commands)

	for i := 0; i < len(commandsByUser); i++ {
		if(commandsByUser[i][0] != "adminDumplogs"){
			wg.Add(1)
			go sendCommands(commandsByUser[i])
		} else {
			wg.Wait()
			wg.Add(1)
			sendCommands(commandsByUser[i])
		}
	}
}
