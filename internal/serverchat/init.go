package serverchat

import (
	"fmt"
	"io"
	"strconv"
)

func StartServer() {

	s, err := NewServer("tcp", ":8080")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer s.CloseListener()

	fmt.Println("Servidor chat escuchando en el puerto 8080...")

	events := make(chan Event, 10)
	go server(events)

	for {
		conn, err := s.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión:", err)
			continue
		}

		e := Event{
			kind:    connection,
			conn:    conn,
			address: conn.RemoteAddress(),
		}

		events <- e

		go handleConnection(conn, events)
	}
}

func server(events chan Event) {

	users := make(map[string]*Client)

	for event := range events {

		switch event.kind {

		case connection:

			newClient := &Client{
				conn:    event.conn,
				in:      make(chan *Outgoing, maxQueue),
				address: event.address,
				done:    make(chan struct{}),
			}

			users[event.address] = newClient

			go worker(newClient, events)

		case disconnection:

			client, ok := users[event.address]

			if !ok || client == nil {
				continue
			}

			close(client.done)
			close(client.in)

			client.conn.CloseConn()

			delete(users, event.address)

		case message:
			for address, client := range users {
				if address == event.address {
					continue
				}
				select {
				case client.in <- &Outgoing{message: event.message}:
				default:
					events <- Event{
						kind:    disconnection,
						address: client.address,
					}
				}

			}

		}
	}
}

func handleConnection(conn *Connection, events chan<- Event) {

	defer conn.CloseConn()

	address := conn.RemoteAddress()

	bodyMessage := "nuevo usuario conectado"

	fmt.Printf("%s -> %s\n", bodyMessage, address)

	headers := make(map[string]string)

	headers["content-type"] = "text/plain"
	headers["size"] = strconv.Itoa(len(bodyMessage))
	headers["sender"] = "servidor"

	payload := conn.BuildHeader(headers)

	events <- Event{
		kind:    message,
		address: address,
		message: payload + bodyMessage,
	}

	for {

		headers, header, err := conn.ReadHeaders()

		if err != nil {

			events <- Event{kind: disconnection, address: address}

			if err == io.EOF {
				println("Usuario desconectado: " + address)
			}

			fmt.Println(err)
			return
		}

		sender, ok := headers["sender"]

		if !ok {
			msg := "falta header 'Sender' en petición"
			fmt.Println(msg)
			conn.SendError(msg, RequestError)
			events <- Event{kind: disconnection, address: address}
			return
		}

		if len(sender) > MaxSenderSize {
			msg := "Valor de header 'sender' de header excede limite tamaño"
			fmt.Println(msg)
			conn.SendError(msg, RequestError)
			events <- Event{kind: disconnection, address: address}
			return
		}

		contentType, ok := headers["content-type"]

		if !ok {
			msg := "Sin header 'content-type'"
			fmt.Println(msg)
			conn.SendError(msg, RequestError)
			events <- Event{kind: disconnection, address: address}
			return
		}

		switch contentType {

		case "text/plain":
			size, err := strconv.Atoi(headers["size"])

			if err != nil {
				msg := "Error en cast de header size: " + err.Error()
				fmt.Println(msg)
				conn.SendError(msg, RequestError)
				continue
			}

			if size < 0 || size > MaxBodySize {
				msg := "Body mensaje excede tamaño limite."
				fmt.Println(msg)
				conn.SendError(msg, RequestError)
				events <- Event{kind: disconnection, address: address}
				return
			}

			buffer := make([]byte, size)

			_, err = io.ReadFull(conn.Reader, buffer)

			if err != nil {
				fmt.Println(err)
				continue
			}

			body := string(buffer)
			header.WriteString(body)

			events <- Event{
				kind:    message,
				message: header.String(),
				address: address,
			}

			fmt.Printf("%s: %s\n", sender, body)

		case "file/notification":
			events <- Event{
				kind:    message,
				message: header.String(),
				address: address,
			}
			fmt.Printf("%s: %s\n", sender, "Notificacion archivo: ")
		default:
			msg := "Content-type desconocido: " + contentType
			fmt.Println(msg)
			conn.SendError(msg, RequestError)
			events <- Event{kind: disconnection, address: address}
			return
		}

		conn.SendOk("Mensaje enviado")

	}

}
