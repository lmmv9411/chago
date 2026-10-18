package client

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolchat"
)

func handleRead(c *Client) {
	defer c.Close()

	for {
		headers, _, err := protocolchat.ReadHeaders(c.conn.reader)
		if err != nil {
			message := strings.TrimSpace(err.Error())
			if errors.Is(err, io.EOF) {
				message = "el servidor cerró la conexión (EOF)"
			} else {
				message = fmt.Sprintf("se interrumpió la conexión con el servidor: %s", message)
			}
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: message,
			})
			return
		}

		size, ok := headers["size"]

		if !ok {
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: "header size no existe.\n",
			})
			return
		}

		n, err := strconv.Atoi(size)
		if err != nil {
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: fmt.Sprintf("Error en cast de header size: %s\n", err),
			})
			return
		}

		if n < 0 {
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: "header: 'size' valor negativo",
			})
			return
		}

		kind, ok := headers["content-type"]

		if !ok {
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: "header content-type no existe.\n",
			})
			return
		}

		switch kind {
		case "text/plain":

			if n < 0 || n > protocolchat.MaxBodySize {
				c.emit(chat.Event{
					Kind: chat.Error,
					Message: "Error: Size-Body declarado en headers" +
						" rango < 0 o superior al límite.",
				})
				return
			}

			buffer := make([]byte, n)

			n, err := io.ReadFull(c.conn.reader, buffer)

			if err != nil {
				c.emit(chat.Event{Kind: chat.Error, Message: err.Error()})
				return
			}

			c.emit(chat.Event{
				Kind: chat.ChatMessage,
				Message: fmt.Sprintf(
					"[%s]: %s", headers["sender"], string(buffer[:n]),
				),
			})

		case "file/notification":
			go func() {
				if err := c.download(headers); err != nil {
					c.emit(chat.Event{Kind: chat.Error, Message: err.Error()})
				}
			}()
		case "response":
			if _, err := response(headers); err != nil {
				c.emit(chat.Event{
					Kind:    chat.Error,
					Message: fmt.Sprintf("%s\n", err.Error()),
				})
				return
			}
		default:
			c.emit(chat.Event{
				Kind:    chat.Error,
				Message: "Error de cabezera 'content-type'= ¡no reconocido!\n",
			})
			return
		}
	}
}
