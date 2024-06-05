package main
import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)
type Request struct {
	verb       string
	path       string
	version    string
	host       string
	user_agent string
}
func parseReq(req string) Request {
	var Req Request
	req_lines := strings.Split(req, "\r\n")
	first_line := strings.Split(req_lines[0], " ")
	Req.verb = first_line[0]
	Req.path = first_line[1]
	Req.version = first_line[2]
	Req.host = strings.Split(req_lines[1], ":")[1]
	Req.user_agent = strings.Split(req_lines[2], ":")[1]
	return Req
}
func handleConnection(c net.Conn) {
	buffer := make([]byte, 5000)
	c.Read(buffer)
	req := string(buffer)
	// Req := parseReq(req)
	// fmt.Println(Req)
	first_line := strings.Split(strings.Split(req, "\r\n")[0], " ")
	if strings.HasPrefix(first_line[1], "/echo/") {
		str := strings.Split(first_line[1], "/")[2]
		c.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(str), str)))
	} else if strings.HasPrefix(first_line[1], "/files/") {
		directory := os.Args[2]
		str := strings.Split(first_line[1], "/")[2]
		data, err := os.ReadFile(directory + str)
		if err != nil {
			c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		} else {
			c.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: " + strconv.Itoa(len(data)) + "\r\n\r\n" + string(data) + "\r\n\r\n"))
		}
	} else if first_line[1] == "/" {
		c.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	}
	if first_line[1] == "/user-agent" {
		// req_agent := strings.Trim(Req.user_agent, " ")
		user_agent := strings.Trim(strings.Split(strings.Split(req, "\r\n")[2], ":")[1], " ")
		c.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(user_agent), user_agent)))
	} else {
		c.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	log.Println("Server running at localhost:4221")
	l, err := net.Listen("tcp", ":4221")
	if err != nil {
		log.Printf("Failed to bind to port 4221")
		os.Exit(69)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Print("Unable to handle connection")
			os.Exit(1)
		}
		go handleConnection(conn)
	}
}