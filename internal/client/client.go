package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolchat"
)

type Client struct {
	conn      *Connection
	name      string
	events    chan chat.Event
	done      chan struct{}
	doneOnce  sync.Once
	closeOnce sync.Once
	closeErr  error
	nextID    atomic.Uint64
}

func (c *Client) Events() <-chan chat.Event {
	return c.events
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func Connect(address string, port string, user string) (*Client, error) {
	user = strings.TrimSpace(user)
	if user == "" {
		return nil, errors.New("el usuario no puede estar vacío")
	}
	if len(user) > protocolchat.MaxSenderSize {
		return nil, fmt.Errorf("el nombre de usuario excede %d bytes", protocolchat.MaxSenderSize)
	}
	if strings.ContainsAny(user, "\r\n") {
		return nil, errors.New("el nombre de usuario contiene caracteres no permitidos")
	}

	connection, err := NewConnection(address, port)
	if err != nil {
		return nil, err
	}

	client := &Client{
		name:   user,
		conn:   connection,
		events: make(chan chat.Event),
		done:   make(chan struct{}),
	}

	go handleRead(client)

	return client, nil
}

func (c *Client) Send(bodyMessage string) error {
	headers := make(map[string]string)

	headers["content-type"] = "text/plain"
	headers["size"] = strconv.Itoa(len(bodyMessage))
	headers["sender"] = c.name

	header := protocolchat.BuildHeader(headers)

	if len(bodyMessage) > protocolchat.MaxBodySize {
		return errors.New("el tamaño del mensaje excede el límite")
	}

	_, err := c.conn.Write([]byte(header + bodyMessage))
	if err != nil {
		return fmt.Errorf("error enviando mensaje: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	c.signalDone()
	c.closeOnce.Do(func() {
		c.closeErr = c.conn.conn.Close()
	})
	return c.closeErr
}

func (c *Client) signalDone() {
	c.doneOnce.Do(func() {
		close(c.done)
	})
}

func (c *Client) emit(event chat.Event) {
	select {
	case c.events <- event:
	case <-c.done:
	}
}

func (c *Client) newTransferID(direction string) string {
	return fmt.Sprintf("%s:%d", direction, c.nextID.Add(1))
}
