package serverchat

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/lmmv9411/chago/internal/protocolchat"
)

const (
	KiB           = 1024
	MaxSenderSize = 64
	MaxBodySize   = KiB
)

type Status int

const (
	Ok Status = iota
	RequestError
	InternalError
)

type Connection struct {
	conn    net.Conn
	Writer  *bufio.Writer
	Reader  *bufio.Reader
	writeMu sync.Mutex
}

type Server struct {
	listener net.Listener
}

func NewServer(network string, address string) (*Server, error) {
	listener, err := net.Listen(network, address)

	if err != nil {
		return nil, err
	}

	return &Server{listener: listener}, nil
}

func (c *Connection) RemoteAddress() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) ReadHeaders() (map[string]string, *strings.Builder, error) {
	return protocolchat.ReadHeaders(c.Reader)
}

func (c *Connection) BuildHeader(headers map[string]string) string {
	return protocolchat.BuildHeader(headers)
}

func (c *Connection) Write(buffer []byte) (int, error) {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	n, err := c.Writer.Write(buffer)
	if err != nil {
		fmt.Println("Error on Write server*: ", err)
		return n, err
	}

	if err := c.Writer.Flush(); err != nil {
		fmt.Println("Error on Flush server*: ", err)
		return n, err
	}

	return n, nil
}

func (s *Server) CloseListener() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

func (s *Connection) CloseConn() error {
	if s.conn != nil {
		return s.conn.Close()
	}
	return nil
}

func (s *Server) Accept() (*Connection, error) {

	conn, err := s.listener.Accept()

	if err != nil {
		return nil, err
	}

	connection := &Connection{
		conn:   conn,
		Writer: bufio.NewWriter(conn),
		Reader: bufio.NewReader(conn),
	}

	return connection, nil
}

func (s *Connection) SendError(msg string, code Status) error {
	return s.send(msg, code)
}

func (s *Connection) SendOk(msg string) error {
	return s.send(msg, Ok)
}

func (s *Connection) send(msg string, code Status) error {
	headers := make(map[string]string)

	headers["content-type"] = "response"
	headers["status"] = strconv.Itoa(int(code))
	headers["size"] = strconv.Itoa(len(msg))
	headers["message"] = msg

	header := protocolchat.BuildHeader(headers)

	_, err := s.Write([]byte(header))

	return err
}
