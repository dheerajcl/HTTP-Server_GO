package main

import (
	//"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

type HTTPRes struct {
	Method    string
	Path      string
	Headers   map[string]string
	BodyMeta  BodyMeta
}

type BodyMeta struct {
	ContentBody string
	ContentLen  int
	ContentType string
}

func NewParser(request string) *HTTPRes {
	lines := strings.Split(request, "\r\n")
	requestLine := strings.Split(lines[0], " ")
	method, path := requestLine[0], requestLine[1]

	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if line == "" {
			break
		}
		headerParts := strings.SplitN(line, ": ", 2)
		if len(headerParts) == 2 {
			headers[headerParts[0]] = headerParts[1]
		}
	}

	return &HTTPRes{
		Method:   method,
		Path:     path,
		Headers:  headers,
		BodyMeta: BodyMeta{},
	}
}

func (res *HTTPRes) FormatRes(statusCode int, statusText string) string {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", statusCode, statusText)
	if res.BodyMeta.ContentType != "" {
		response += fmt.Sprintf("Content-Type: %s\r\n", res.BodyMeta.ContentType)
	}
	if res.BodyMeta.ContentLen != 0 {
		response += fmt.Sprintf("Content-Length: %d\r\n", res.BodyMeta.ContentLen)
	}
	response += "\r\n"
	if res.BodyMeta.ContentBody != "" {
		response += res.BodyMeta.ContentBody
	}
	return response
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
		// concurrent
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	for {
		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("could not read buffer")
			os.Exit(1)
		}
		received := string(buff[:n])
		parser := NewParser(received)
		httpRes := parser
		var res string
		fmt.Println(httpRes)
		if httpRes.Path == "/echo" || httpRes.Path == "/" {
			res = httpRes.FormatRes(200, "OK")
		} else if httpRes.Path == "/user-agent" {
			bodyRes := httpRes.Headers["User-Agent"]
			httpRes.BodyMeta.ContentBody = bodyRes
			httpRes.BodyMeta.ContentLen = len(bodyRes)
			httpRes.BodyMeta.ContentType = "text/plain"
			res = httpRes.FormatRes(200, "OK")
		} else {
			res = "HTTP/1.1 404 Not Found\r\n\r\n"
		}
		conn.Write([]byte(res))
	}
}
