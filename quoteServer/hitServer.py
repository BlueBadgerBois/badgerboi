import socket

s = socket.socket(
    socket.AF_INET, socket.SOCK_STREAM)

s.connect(("localhost", 3333))
s.send("ABC, edgar")
data = s.recv(1024)
print data
s.close()
