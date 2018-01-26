#client code handles sending POST requests the right way

import requests

command = raw_input("Enter a command")
command = command.split(",")
[i.strip() for i in command]

print(command)

if command[0] == 'ADD':
	req = requests.post("localhost:8082/add", data = {"uID": command[1], "amount": command[2]})	

print(req.text)