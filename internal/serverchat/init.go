package serverchat

import (
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
)

func StartServer() error {

	server, err := NewServer("tcp", ":8080")

	if err != nil {
		return fmt.Errorf("error al iniciar el servidor de chat: %w", err)
	}

	defer server.CloseListener()

	fmt.Println("Servidor chat escuchando en el puerto 8080...")

	events := make(chan Event, 100)
	go eventLoop(events)

	var backoff time.Duration
	for {

		conn, err := server.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				if backoff == 0 {
					backoff = 5 * time.Millisecond
				} else {
					backoff *= 2
				}
				if backoff > time.Second {
					backoff = time.Second
				}
				fmt.Println("Error temporal al aceptar conexión:", err)
				time.Sleep(backoff)
				continue
			}
			return fmt.Errorf("error al aceptar conexión de chat: %w", err)
		}
		backoff = 0

		e := Event{
			kind:    connection,
			conn:    conn,
			address: conn.RemoteAddress(),
		}

		events <- e

		go handleConnection(conn, events)
	}
}

const maxClients = 256

func eventLoop(events chan Event) {

	users := make(map[string]*Client)

	disconnect := func(address string) {
		client, ok := users[address]
		if !ok || client == nil {
			return
		}

		close(client.done)
		close(client.in)
		client.conn.CloseConn()

		delete(users, address)
	}

	broadcast := func(event Event) {
		for address, client := range users {
			if address == event.address {
				continue
			}
			select {
			case client.in <- &Outgoing{message: event.message}:
			default:
				fmt.Printf("Cola llena: desconectando cliente %s\n", address)
				disconnect(address)
			}
		}
	}

	for event := range events {

		switch event.kind {

		case connection:
			if len(users) >= maxClients {
				if err := event.conn.SendError("servidor lleno", RequestError); err != nil {
					fmt.Println("Error rechazando cliente:", err)
				}
				if err := event.conn.CloseConn(); err != nil {
					fmt.Println("Error cerrando cliente rechazado:", err)
				}
				break
			}

			newClient := &Client{
				conn:    event.conn,
				in:      make(chan *Outgoing, maxQueue),
				address: event.address,
				done:    make(chan struct{}),
			}

			users[event.address] = newClient

			go worker(newClient, events)

			body := "nuevo usuario conectado"
			headers := map[string]string{
				"content-type": "text/plain",
				"size":         strconv.Itoa(len(body)),
				"sender":       "servidor",
			}
			broadcast(Event{
				kind:    message,
				address: event.address,
				message: protocolchat.BuildHeader(headers) + body,
			})

		case disconnection:

			disconnect(event.address)

		case message:
			broadcast(event)

		}
	}
}

func handleConnection(conn *Connection, events chan<- Event) {

	address := conn.RemoteAddress()

	reject := func(msg string, code Status) {
		fmt.Println(msg)
		if err := conn.SendError(msg, code); err != nil {
			fmt.Println("Error enviando respuesta de error:", err)
		}
		events <- Event{kind: disconnection, address: address}
	}

	for {

		headers, header, err := conn.ReadHeaders()

		if err != nil {

			events <- Event{kind: disconnection, address: address}

			if err == io.EOF {
				println("Usuario desconectado: " + address)
			} else {
				fmt.Printf("Error leyendo de %s: %v\n", address, err)
			}

			return
		}

		sender, ok := headers["sender"]

		if !ok {
			reject("falta header 'sender' en petición", RequestError)
			return
		}

		if len(sender) > protocolchat.MaxSenderSize || containsControl(sender) {
			reject("valor de header 'sender' inválido", RequestError)
			return
		}

		contentType, ok := headers["content-type"]

		if !ok {
			reject("sin header 'content-type'", RequestError)
			return
		}

		switch contentType {

		case "text/plain":
			size, err := strconv.Atoi(headers["size"])

			if err != nil {
				reject("error en header size: "+err.Error(), RequestError)
				return
			}

			if size < 0 || size > protocolchat.MaxBodySize {
				reject("body del mensaje excede el límite", RequestError)
				return
			}

			buffer := make([]byte, size)

			_, err = io.ReadFull(conn.Reader, buffer)

			if err != nil {
				reject("error leyendo el body: "+err.Error(), InternalError)
				return
			}

			body := sanitizeMessage(string(buffer))
			headers["size"] = strconv.Itoa(len(body))

			events <- Event{
				kind:    message,
				message: protocolchat.BuildHeader(headers) + body,
				address: address,
			}

			//fmt.Printf("%s: %s\n", sender, body)

		case "file/notification":
			if err := validateFileNotification(headers); err != nil {
				msg := "Notificación de archivo inválida: " + err.Error()
				reject(msg, RequestError)
				return
			}
			events <- Event{
				kind:    message,
				message: header.String(),
				address: address,
			}
			fmt.Printf("%s: %s\n", sender, "Notificacion archivo: ")
		default:
			reject("content-type desconocido: "+contentType, RequestError)
			return
		}

		if err := conn.SendOk("Mensaje enviado"); err != nil {
			fmt.Printf("Error enviando respuesta a %s: %v\n", address, err)
			events <- Event{kind: disconnection, address: address}
			return
		}

	}
}

func validateFileNotification(headers map[string]string) error {
	filename, ok := headers["filename"]
	if !ok || filename == "" || filename == "." || filename == ".." ||
		filepath.Base(filename) != filename || strings.ContainsAny(filename, "/\\:") ||
		containsControl(filename) {
		return fmt.Errorf("nombre de archivo inválido")
	}

	size, err := strconv.ParseInt(headers["size"], 10, 64)
	if err != nil {
		return fmt.Errorf("tamaño inválido: %w", err)
	}
	if size < 0 || size > protocolfile.MaxBodySize {
		return fmt.Errorf("tamaño fuera del límite permitido")
	}

	return nil
}

func containsControl(value string) bool {
	return strings.IndexFunc(value, unicode.IsControl) >= 0
}

func sanitizeMessage(message string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' {
			return ' '
		}
		return r
	}, message)
}
