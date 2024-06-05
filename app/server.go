package main
import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)
const (
	CRLF = "\r\n"
)
type HTTPRequest struct {
	Method      string
	Target      string
	HTTPVersion string
	Headers     map[string]string
	Body        []byte
}
func (r HTTPRequest) String() string {
	return fmt.Sprintf("%s %s %s %v", r.Method, r.Target, r.HTTPVersion, r.Body)
}
func byteReader(channel chan []byte, connection net.Conn) error {
	buffer := make([]byte, 0)
	for {
		tmp := make([]byte, 1024)
		n, err := connection.Read(tmp)
		if err != nil && err != io.EOF {
			close(channel)
			return err
		}
		buffer = append(buffer, tmp[:n]...)
		splitBuffer := bytes.Split(buffer, []byte(CRLF))
		for _, line := range splitBuffer[:len(splitBuffer)-1] {
			channel <- line
		}
		buffer = splitBuffer[len(splitBuffer)-1]
		if n <= len(tmp) || err == io.EOF {
			break
		}
	}
	channel <- buffer
	close(channel)
	return nil
}
func parseRequest(connection net.Conn) (HTTPRequest, error) {
	lines := make(chan []byte)
	go byteReader(lines, connection)
	requestLineValues := bytes.Split(<-lines, []byte(" "))
	if len(requestLineValues) != 3 {
		return HTTPRequest{}, fmt.Errorf("invalid request line")
	}
	method := string(requestLineValues[0])
	target := string(requestLineValues[1])
	httpVersion := string(requestLineValues[2])
	headers := make(map[string]string)
	for {
		headerLine := <-lines
		if len(headerLine) == 0 {
			break
		}
		headerLineValues := bytes.Split(headerLine, []byte(": "))
		headers[string(headerLineValues[0])] = string(headerLineValues[1])
	}
	body := <-lines
	return HTTPRequest{
		Method:      method,
		Target:      target,
		HTTPVersion: httpVersion,
		Headers:     headers,
		Body:        body,
	}, nil
}
type HTTPResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte
}
func (r HTTPResponse) String() string {
	response := fmt.Sprintf("HTTP/1.1 %d %s\r\n", r.StatusCode, http.StatusText(r.StatusCode))
	for k, v := range r.Headers {
		response += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	response += "\r\n" + string(r.Body)
	return response
}
func (r HTTPResponse) Write(connection net.Conn) error {
	_, err := connection.Write([]byte(r.String()))
	return err
}
func textResponse(statusCode int, body string) HTTPResponse {
	contentLength := len(body)
	return HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "text/plain", "Content-Length": fmt.Sprintf("%d", contentLength)},
		Body:       []byte(body),
	}
}
func fileResponse(statusCode int, file *os.File) HTTPResponse {
	stat, err := file.Stat()
	if err != nil {
		return textResponse(500, err.Error())
	}
	contentLength := stat.Size()
	contents, err := io.ReadAll(file)
	if err != nil {
		return textResponse(500, err.Error())
	}
	return HTTPResponse{
		StatusCode: statusCode,
		Headers:    map[string]string{"Content-Type": "application/octet-stream", "Content-Length": fmt.Sprintf("%d", contentLength)},
		Body:       contents,
	}
}
type HTTPHandler func(HTTPRequest) HTTPResponse
type HTTPRouter struct {
	routes map[string]HTTPHandler
}
func NewHTTPRouter() HTTPRouter {
	return HTTPRouter{
		routes: make(map[string]HTTPHandler),
	}
}
func (s *HTTPRouter) AddEndpoint(method string, prefix string, handler HTTPHandler) {
	s.routes[method+prefix] = handler
}
func (s *HTTPRouter) GetHandler(method string, prefix string) HTTPHandler {
	if handler, ok := s.routes[method+prefix]; ok {
		return handler
	}
	for routeInfo := range s.routes {
		if strings.HasSuffix(routeInfo, "*") && strings.HasPrefix(method+prefix, routeInfo[:len(routeInfo)-1]) {
			return s.routes[routeInfo]
		}
	}
	return nil
}
func (s HTTPRouter) handleConnection(connection net.Conn) error {
	defer connection.Close()
	request, err := parseRequest(connection)
	if err != nil {
		fmt.Println("Failed to parse request: ", err.Error())
		return err
	}
	handler := s.GetHandler(request.Method, request.Target)
	var response HTTPResponse
	if handler == nil {
		response = HTTPResponse{StatusCode: 404}
	} else {
		response = handler(request)
	}
	if err := response.Write(connection); err != nil {
		fmt.Println("Error writing response: ", err.Error())
		return err
	}
	fmt.Println("Processed request: ", request.String())
	return nil
}
func main() {
	var directory string
	flag.StringVar(&directory, "directory", "", "Directory to serve")
	flag.Parse()
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	router := NewHTTPRouter()
	router.AddEndpoint("GET", "/", func(request HTTPRequest) HTTPResponse {
		return HTTPResponse{StatusCode: 200}
	})
	router.AddEndpoint("GET", "/echo/*", func(request HTTPRequest) HTTPResponse {
		data := strings.TrimPrefix(request.Target, "/echo/")
		response := textResponse(200, data)
		encoding, ok := request.Headers["Accept-Encoding"]
		if ok && encoding == "gzip" {
			response.Headers["Content-Encoding"] = encoding
		}
		return response
	})
	router.AddEndpoint("GET", "/user-agent", func(request HTTPRequest) HTTPResponse {
		value, ok := request.Headers["User-Agent"]
		if !ok {
			return HTTPResponse{StatusCode: 400}
		}
		return textResponse(200, value)
	})
	router.AddEndpoint("GET", "/files/*", func(request HTTPRequest) HTTPResponse {
		filePath := strings.TrimPrefix(request.Target, "/files/")
		file, err := os.Open(filepath.Join(directory, filePath))
		if err != nil {
			return HTTPResponse{StatusCode: 404}
		}
		defer file.Close()
		return fileResponse(200, file)
	})
	router.AddEndpoint("POST", "/files/*", func(request HTTPRequest) HTTPResponse {
		filePath := strings.TrimPrefix(request.Target, "/files/")
		file, err := os.Create(filepath.Join(directory, filePath))
		if err != nil {
			return HTTPResponse{StatusCode: 500}
		}
		defer file.Close()
		_, err = file.Write(request.Body)
		if err != nil {
			return HTTPResponse{StatusCode: 500}
		}
		return HTTPResponse{StatusCode: 201}
	})
	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
		}
		fmt.Println("Connected: ", connection.RemoteAddr())
		go router.handleConnection(connection)
	}
}