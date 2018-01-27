/* Workload Generator */

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"net/http"
	"net/url"
)

const serverUrl = "http://127.0.0.1:8082"

/*
* Returns a two dimentional array, outer array contains the users,
* inner array contains the username in index [0] and his commands 
* in the remaining indexes.
*/
func divideCommandsByUser (commands []string) [][]string {
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
	var command []string

	for i := 1; i < len(commands); i++ {
		command = strings.Split(commands[i], ",")
		fmt.Println("----------------------------------------------")
		fmt.Println(command)

		if command[0] == "ADD" {
			data := url.Values{}
			data.Set("username", command[1])
			data.Add("amount", command[2])
			req, err := http.NewRequest("POST", serverUrl + "/add", strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Println("Error making a new request:")
				fmt.Println(err)
			}
			sendRequest(req)
		}

		if command[0] == "QUOTE" {
			data := url.Values{}
			data.Set("username", command[1])
			data.Add("stocksym", command[2])
			req, err := http.NewRequest("POST", serverUrl + "/quote", strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Println("Error making a new request:")
				fmt.Println(err)
			}
			sendRequest(req)
		}

		if command[0] == "DISPLAY_SUMMARY" {
			data := url.Values{}
			data.Set("username", command[1])
			req, err := http.NewRequest("POST", serverUrl + "/summary", strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Println("Error making a new request:")
				fmt.Println(err)
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
				fmt.Println("Error making a new request:")
				fmt.Println(err)
			}
			sendRequest(req)
		}

		if command[0] == "SET_BUY_TRIGGER" {
			data := url.Values{}
			data.Set("username", command[1])
			data.Add("stockSymbol", command[2])
			data.Add("buyAmount", command[3])
			data.Add("threshold", command[4])
			req, err := http.NewRequest("POST", serverUrl + "/buy", strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Println("Error making a new request:")
				fmt.Println(err)
			}
			sendRequest(req)
		}

		if command[0] == "DUMPLOG" {
			data := url.Values{}
			data.Set("username", command[1])
			data.Add("outfile", command[2])
			req, err := http.NewRequest("POST", serverUrl + "/dumplog", strings.NewReader(data.Encode()))
			if err != nil {
				fmt.Println("Error making a new request:")
				fmt.Println(err)
			}
			sendRequest(req)
		}
	}
}

/*
* Handles sending a request and printing the response
*/
func sendRequest (req *http.Request) {
	fmt.Println("Sending request...")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending http request:")
		panic(err)
	}
	defer resp.Body.Close()
	fmt.Println("Response Status:", resp.Status)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response Body:", string(body))
}

func main (){
	if len(os.Args) < 2 {
		fmt.Println("Filename not provided")
		return
	}

	fileData, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println("Error reading file ", os.Args[1])
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
		go sendCommands(commandsByUser[i])
	}
}