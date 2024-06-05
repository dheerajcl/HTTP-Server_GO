package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	
	buffer:= make([]byte, 1024)

	conn.Read(buffer)

	splitHeader:=strings.Split(string(buffer), "\r\n")
	splitRequestLine:=strings.Split(splitHeader[0], " ")

	if splitRequestLine[1] == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else if strings.Split(splitRequestLine[1], "/")[1] == "echo" {
		requestBody:=strings.Split(splitRequestLine[1], "/")[2]
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: "+fmt.Sprint(len(requestBody))+"\r\n\r\n"+requestBody))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
