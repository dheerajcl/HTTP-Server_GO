
package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
)
func main() {
	fmt.Println("Logs from your program will appear here!")
	directoryPtr := flag.String("directory", "", "a string")
	flag.Parse()
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go func() {
			buff := make([]byte, 2048)
			_, err = conn.Read(buff)
			if err != nil {
				fmt.Println("Error reading request: ", err.Error())
				os.Exit(1)
			}
			request := string(buff)
			requestMethod := strings.Split(request, " ")[0]
			requestTarget := strings.Split(request, " ")[1]
			headers := strings.Split(request, "\r\n")[1:]
			var response string
			if requestTarget == "/" {
				response = "HTTP/1.1 200 OK\r\n\r\n"
			} else if strings.HasPrefix(requestTarget, "/echo") {
				echoString := strings.Split(requestTarget, "/echo/")[1]
				acceptEncodingIndex := slices.IndexFunc(headers, func(h string) bool { return strings.Contains(h, "Accept-Encoding: ") })
				if acceptEncodingIndex == -1 {
					response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoString), echoString)
				} else {
					acceptEncoding := strings.TrimPrefix(headers[acceptEncodingIndex], "Accept-Encoding: ")
					if strings.Contains(acceptEncoding, "gzip") {
						var b bytes.Buffer
						enc := gzip.NewWriter(&b)
						enc.Write([]byte(echoString))
						enc.Close()
						response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(b.String()), b.String())
					} else {
						response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(echoString), echoString)
					}
				}
			} else if strings.HasPrefix(requestTarget, "/user-agent") {
				userAgentIndex := slices.IndexFunc(headers, func(h string) bool { return strings.Contains(h, "User-Agent: ") })
				userAgent := strings.TrimPrefix(headers[userAgentIndex], "User-Agent: ")
				response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", len(userAgent), userAgent)
			} else if strings.HasPrefix(requestTarget, "/files") {
				fileName := strings.TrimPrefix(requestTarget, "/files/")
				if requestMethod == "GET" {
					contents, err := os.ReadFile(*directoryPtr + "/" + fileName)
					if err != nil {
						response = "HTTP/1.1 404 Not Found\r\n\r\n"
					} else {
						response = fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(contents), contents)
					}
				} else if requestMethod == "POST" {
					body := headers[len(headers)-1]
					bodyBytes := bytes.Trim([]byte(body), "\x00")
					fmt.Print(body)
					err := os.WriteFile(*directoryPtr+"/"+fileName, bodyBytes, 0644)
					if err != nil {
						response = ""
					} else {
						response = "HTTP/1.1 201 Created\r\n\r\n"
					}
				} else {
					response = ""
				}
			} else {
				response = "HTTP/1.1 404 Not Found\r\n\r\n"
			}
			conn.Write([]byte(response))
			conn.Close()
		}()
	}
}