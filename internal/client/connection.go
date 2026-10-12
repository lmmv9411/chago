package client

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"time"
)

type Connection struct {
	conn    net.Conn
	address string
	writer  *bufio.Writer
	reader  *bufio.Reader
	mu      sync.Mutex
}

func NewConnection(address string, port string) (connection *Connection, err error) {

	if address == "" || port == "" {
		return nil, errors.New("campos vacíos[address || port]")
	}

	if net.ParseIP(address) == nil {
		return nil, errors.New("dirección ip inválida")
	}

	conn, err := net.DialTimeout("tcp", address+port, 5*time.Second)

	if err != nil {
		return nil, err
	}

	return &Connection{
		conn:    conn,
		address: address,
		writer:  bufio.NewWriter(conn),
		reader:  bufio.NewReader(conn),
	}, nil
}

func (c *Connection) Write(p []byte) (n int, err error) {
	defer c.mu.Unlock()
	c.mu.Lock()
	return c.conn.Write(p)
}
