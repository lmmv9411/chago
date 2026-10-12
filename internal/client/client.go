package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/events"
	"github.com/lmmv9411/chago/internal/protocolchat"
)

type Client struct {
	conn   *Connection
	name   string
	events chan events.Event
}

func (c *Client) Events() chan events.Event {
	return c.events
}

func Connect(address string, port string, user string) (*Client, error) {

	if user == "" {
		return nil, errors.New("Usuario no puede esar vacío")
	}

	connection, err := NewConnection(address, port)

	if err != nil {
		return nil, err
	}

	client := &Client{
		name:   user,
		conn:   connection,
		events: make(chan events.Event),
	}

	go handleRead(client)

	return client, nil

}
func (c *Client) Send(bodyMessage string) error {
	headers := make(map[string]string)

	if after, ok := strings.CutPrefix(bodyMessage, "/file"); ok {

		filePath := strings.TrimSpace(after)

		if filePath == "" {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: "Escribir la ruta del archivo",
			}
			return nil
		}

		go func() {
			if err := c.upload(filePath); err != nil {
				c.events <- events.Event{Kind: events.Error, Message: err.Error()}
			}
		}()

		return nil
	}

	headers["content-type"] = "text/plain"
	headers["size"] = strconv.Itoa(len(bodyMessage))
	headers["sender"] = c.name

	header := protocolchat.BuildHeader(headers)

	_, err := c.conn.Write([]byte(header + bodyMessage))

	if err != nil {
		c.events <- events.Event{
			Kind:    events.Error,
			Message: fmt.Sprintf("Error enviando mensaje: %s\n", err),
		}
		return err
	}

	return nil
}

func (c *Client) Close() error {
	if err := c.conn.conn.Close(); err != nil {
		return err
	}
	return nil
}
