package client

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/ui"
	"github.com/lmmv9411/chago/internal/ui/chat"
)

func StartClient() {

	data := ui.LoginMain()

	if data == nil {
		return
	}

	if data.Address == "" {
		data.Address = "192.168.1.33"
	}

	if data.User == "" {
		fmt.Println("Usuario no puede esar vacío")
		return
	}

	server, err := NewConnection(data.Address, ":8080")

	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}

	defer server.conn.Close()

	client := NewClient(data.User, server)

	fmt.Println("Conectado al servidor")

	go handleWrite(client)
	go handleRead(client)

	chat.Run(client.incoming, client.events)
}

func handleRead(c *Client) {

	for {
		headers, _, err := protocolchat.ReadHeaders(c.conn.reader)
		if err != nil {
			c.events <- chat.ErrorChatMsg(fmt.Sprintf("%s\n", err.Error()))
			return
		}

		size, ok := headers["size"]

		if !ok {
			c.events <- chat.ErrorChatMsg("header size no existe.\n")
			return
		}

		n, err := strconv.Atoi(size)
		if err != nil {
			c.events <- chat.
				ErrorChatMsg(
					fmt.Sprintf("Error en cast de header size: %s\n", err),
				)
			return
		}

		if n < 0 {
			c.events <- chat.
				ErrorChatMsg("header: 'size' valor negativo")
			return
		}

		kind, ok := headers["content-type"]

		if !ok {
			c.events <- chat.ErrorChatMsg("header content-type no existe.\n")
			return
		}

		switch kind {
		case "text/plain":

			if n < 0 || n > protocolchat.MaxBodySize {
				c.events <- chat.ErrorChatMsg(
					"Error: Size-Body declarado en headers" +
						" rango < 0 o superior al límite.")
				return
			}

			buffer := make([]byte, n)

			n, err := io.ReadFull(c.conn.reader, buffer)

			if err != nil {
				c.events <- chat.ErrorChatMsg(fmt.Sprintf("%s", err.Error()))
				return
			}

			c.events <- chat.
				IncomingChatMsg(
					fmt.Sprintf("[%s]: %s",
						headers["sender"],
						string(buffer[:n])),
				)

		case "file/notification":
			go downloadFile(headers, c)
		case "response":
			if _, err := response(headers); err != nil {
				c.events <- chat.ErrorChatMsg(fmt.Sprintf("%s\n", err.Error()))
				return
				// } else {
				// 	c.events <- chat.InconmingChatMsg(fmt.Sprintf("%s\n", *msg))
			}
		default:
			c.events <- chat.
				ErrorChatMsg(
					"Error de cabezera 'content-type'= ¡no reconocido!\n",
				)
			return
		}
	}
}

func handleWrite(c *Client) {

	headers := make(map[string]string)

	for msg := range c.incoming {

		bodyMessage := msg

		if after, ok := strings.CutPrefix(bodyMessage, "/file"); ok {

			filePath := strings.TrimSpace(after)

			if filePath == "" {
				c.events <- chat.ErrorChatMsg("Escribir la ruta del archivo")
				continue
			}

			go sendToServer(filePath, c)

			continue
		}

		headers["content-type"] = "text/plain"
		headers["size"] = strconv.Itoa(len(bodyMessage))
		headers["sender"] = c.name

		header := protocolchat.BuildHeader(headers)

		_, err := c.conn.Write([]byte(header + bodyMessage))

		if err != nil {
			c.events <- chat.ErrorChatMsg(fmt.Sprintf("Error enviando mensaje: %s\n", err))
			return
		}

	}
}
