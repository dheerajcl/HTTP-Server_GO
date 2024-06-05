package main
import (
	"fmt"
	"net"
	"os"
	"strings"
)
type Request struct {
	method  string
	path    string
	version string
	headers map[string]string
}
func parseRequest(buffer []byte) Request {
	bufferString := string(buffer)
	lines := strings.Split(bufferString, "\r\n")
	requestLine := strings.Split(lines[0], " ")
	headers := make(map[string]string)
	for i := 1; i < len(lines); i++ {
		fmt.Println("setting header", lines[i])
		header := strings.Split(lines[i], ": ")
		if len(header) > 1 {
			headers[strings.ToLower(header[0])] = header[1]
		}
	}
	return Request{
		method:  requestLine[0],
		path:    requestLine[1],
		version: requestLine[2],
		headers: headers,
	}
}
func buildResponse(body string, statusLine string) string {
	return fmt.Sprintf("HTTP/1.1 %s\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", statusLine, len(body), body)
}
func handleRequest(request Request, conn net.Conn) {
	if strings.HasPrefix(request.path, "/echo/") {
		suffix := strings.Split(request.path[6:], "/")[0]
		fmt.Println("Suffix", strings.Split(suffix, "/"))
		response := buildResponse(suffix, "200 OK")
		conn.Write([]byte(response))
	} else if request.path == "/user-agent" {
		userAgentHeaderValue, _ := request.headers["user-agent"]
		response := buildResponse(userAgentHeaderValue, "200 OK")
		conn.Write([]byte(response))
	} else if request.path == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\nHello, World!"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
	defer conn.Close()
}
func main() {
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
		defer l.Close()
		buffer := make([]byte, 1024)
		_, err = conn.Read(buffer)
		request := parseRequest(buffer)
		go handleRequest(request, conn)
	}
}