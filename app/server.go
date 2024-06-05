package main
import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
)
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}
func handleConn(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error handling request: ", err.Error())
		os.Exit(1)
	}
	r := strings.Split(string(buf), "\r\n")
	m := strings.Split(r[0], " ")[0]
	p := strings.Split(r[0], " ")[1]
	var response string
	if m == "GET" && p == "/" {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if m == "GET" && p[0:6] == "/echo/" {
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:%d\r\n\r\n%s", len(p[6:]), p[6:])
	} else if m == "GET" && p == "/user-agent" {
		ua := strings.Split(r[2], " ")[1]
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length:%d\r\n\r\n%s", len(ua), ua)
	} else if m == "GET" && p[0:7] == "/files/" {
		dir := os.Args[2]
		content, err := os.ReadFile(path.Join(dir, p[7:]))
		if err != nil {
			response = "HTTP/1.1 404 Not Found\r\n\r\n"
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(content), string(content))
		}
	} else if m == "POST" && p[0:7] == "/files/" {
		content := strings.Trim(r[len(r)-1], "\x00")
		dir := os.Args[2]
		_ = os.WriteFile(path.Join(dir, p[7:]), []byte(content), 0644)
		response = "HTTP/1.1 201 Created\r\n\r\n"
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	conn.Write([]byte(response))
}