
package main
import (
	"bytes"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
)
const CRLF = "\r\n"
type StatusLine struct {
	method  string
	path    string
	version string
}
type Headers = map[string]string
type Body = string
type Request struct {
	status  StatusLine
	headers Headers
	body    Body
}
func parseRequest(buffer []byte) Request {
	buffer = bytes.Trim(buffer, "\x00")
	stringBuffer := string(buffer)
	return Request{
		status:  makeStatusLine(stringBuffer),
		headers: makeHeaders(stringBuffer),
		body:    makeBody(stringBuffer),
	}
}
func makeStatusLine(buffer string) StatusLine {
	statusLineIndex := strings.Index(buffer, CRLF)
	statusLine := buffer[:statusLineIndex]
	stringStatusLine := strings.Fields(statusLine)
	return StatusLine{
		method:  stringStatusLine[0],
		path:    stringStatusLine[1],
		version: stringStatusLine[2],
	}
}
func makeHeaders(buffer string) Headers {
	headers := make(map[string]string)
	lines := strings.Split(buffer, CRLF)
	for _, line := range lines {
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	return headers
}
func makeBody(buffer string) Body {
	bodyIndex := strings.LastIndex(buffer, CRLF)
	body := buffer[bodyIndex:]
	body = body[2:]
	return body
}
func handleConnection(connection net.Conn) {
	buffer := make([]byte, 1024)
	connection.Read(buffer)
	request := parseRequest(buffer)
	homePattern := "/"
	echoPattern := regexp.MustCompile(`^/echo/[a-zA-Z0-9]+$`)
	userAgentPattern := "/user-agent"
	filenamePattern := regexp.MustCompile(`^/files/[a-zA-Z0-9_]+$`)
	var response string
	if echoPattern.MatchString(request.status.path) {
		pathComponents := strings.Split(request.status.path, "/")
		path := pathComponents[2]
		value, ok := request.headers["Accept-Encoding"]
		if ok {
			if value == "gzip" {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path), path)
			} else if strings.Contains(value, "gzip") {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path), path)
			} else {
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path), path)
			}
		} else {
			response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(path), path)
		}
	} else if userAgentPattern == request.status.path {
		userAgentValue := request.headers["User-Agent"]
		response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgentValue), userAgentValue)
	} else if homePattern == request.status.path {
		response = "HTTP/1.1 200 OK\r\n\r\n"
	} else if filenamePattern.MatchString(request.status.path) {
		pathComponents := strings.Split(request.status.path, "/")
		filename := pathComponents[2]
		directory := os.Args[2]
		if request.status.method == "GET" {
			data, err := os.ReadFile(directory + filename)
			if err != nil {
				fmt.Println("cannot find")
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			} else {
				fmt.Println("fined")
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(data), data)
			}
		} else {
			os.WriteFile(directory+filename, []byte(request.body), os.ModeTemporary)
			response = "HTTP/1.1 201 Created\r\n\r\n"
		}
	} else {
		response = "HTTP/1.1 404 Not Found\r\n\r\n"
	}
	byteResponse := []byte(response)
	connection.Write(byteResponse)
	connection.Close()
}
func main() {
	fmt.Println("Server was started")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		connection, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(connection)
	}
}