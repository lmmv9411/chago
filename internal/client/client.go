package client

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolchat"
)

type Client struct {
	conn   *Connection
	name   string
	events chan chat.Event
}

func (c *Client) Events() <-chan chat.Event {
	return c.events
}

func Connect(address string, port string, user string) (*Client, error) {
	if user == "" {
		return nil, errors.New("usuario no puede esar vacío")
	}

	connection, err := NewConnection(address, port)
	if err != nil {
		return nil, err
	}

	client := &Client{
		name:   user,
		conn:   connection,
		events: make(chan chat.Event),
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
	if err := c.conn.conn.Close(); err != nil {
		return err
	}
	return nil
}
