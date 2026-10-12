package client

import (
	"fmt"
	"io"
	"strconv"

	"github.com/lmmv9411/chago/internal/events"
	"github.com/lmmv9411/chago/internal/protocolchat"
)

func handleRead(c *Client) {

	for {
		headers, _, err := protocolchat.ReadHeaders(c.conn.reader)
		if err != nil {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: fmt.Sprintf("%s\n", err.Error()),
			}
			return
		}

		size, ok := headers["size"]

		if !ok {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: "header size no existe.\n",
			}
			return
		}

		n, err := strconv.Atoi(size)
		if err != nil {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: fmt.Sprintf("Error en cast de header size: %s\n", err),
			}
			return
		}

		if n < 0 {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: "header: 'size' valor negativo",
			}
			return
		}

		kind, ok := headers["content-type"]

		if !ok {
			c.events <- events.Event{
				Kind:    events.Error,
				Message: "header content-type no existe.\n"}
			return
		}

		switch kind {
		case "text/plain":

			if n < 0 || n > protocolchat.MaxBodySize {
				c.events <- events.Event{
					Kind: events.Error,
					Message: "Error: Size-Body declarado en headers" +
						" rango < 0 o superior al límite.",
				}
				return
			}

			buffer := make([]byte, n)

			n, err := io.ReadFull(c.conn.reader, buffer)

			if err != nil {
				c.events <- events.Event{Kind: events.Error, Message: err.Error()}
				return
			}

			c.events <- events.Event{
				Kind: events.ChatMessage,
				Message: fmt.Sprintf(
					"[%s]: %s", headers["sender"], string(buffer[:n]),
				),
			}

		case "file/notification":
			go func() {
				if err := c.download(headers); err != nil {
					c.events <- events.Event{Kind: events.Error, Message: err.Error()}
				}
			}()
		case "response":
			if _, err := response(headers); err != nil {
				c.events <- events.Event{
					Kind:    events.Error,
					Message: fmt.Sprintf("%s\n", err.Error()),
				}
				return
				// } else {
				// 	c.events <- chat.InconmingChatMsg(fmt.Sprintf("%s\n", *msg))
			}
		default:
			c.events <- events.Event{
				Kind:    events.Error,
				Message: "Error de cabezera 'content-type'= ¡no reconocido!\n",
			}
			return
		}
	}
}
