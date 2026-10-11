package client

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Connection struct {
	conn    net.Conn
	address string
	writer  *bufio.Writer
	reader  *bufio.Reader
	mu      sync.Mutex
}

type Client struct {
	conn     *Connection
	name     string
	incoming chan string
	events   chan tea.Msg
}

func NewClient(name string, conn *Connection) *Client {
	return &Client{
		name:     name,
		conn:     conn,
		incoming: make(chan string),
		events:   make(chan tea.Msg),
	}
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
