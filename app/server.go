package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
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

	defer conn.Close()

	request, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request. ", err.Error())
		return
	}
	
	var res string
	path:=request.URL.Path

	if path == "/" {
		res = "HTTP/1.1 200 OK\r\n\r\n"
	} else if path == "/user-agent" {
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(request.UserAgent()), request.UserAgent())
	} else if path[0:6] == "/echo/" {
		echo := path[6:]
		res = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echo), echo)
	} else {
		res = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(res))
}
