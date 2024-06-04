package main

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"os"
)

func Handler(conn net.Conn) {
	defer conn.Close()
	reader, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		fmt.Println("Error reading request. ", err.Error())
		return
	}
	fmt.Printf("Request: %s %s\n", reader.Method, reader.URL.Path)
	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))

	if reader.URL.Path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		return
	}
	conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
}

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	fmt.Println("Listening on port 4221")
	defer l.Close()
	conn, err := l.Accept()
	if err != nil {
		fmt.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	Handler(conn)
}
