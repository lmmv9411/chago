package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
)

func StartServer() {

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Servidor escuchando en el puerto 8080...")

	events := make(chan Event, 10)
	go server(events)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión:", err)
			continue
		}

		e := Event{
			kind:    connection,
			conn:    conn,
			address: conn.RemoteAddr().String()}

		events <- e

		go handleConnection(conn, events)
	}
}

func server(events chan Event) {

	users := make(map[string]*Client)

	for event := range events {
		switch event.kind {
		case connection:
			users[event.address] = &Client{
				conn:    event.conn,
				out:     make(chan Outgoing, 10),
				address: event.address,
			}
			go worker(users[event.address], events)

		case disconnection:
			client, ok := users[event.address]
			if !ok || client == nil {
				continue
			}
			close(client.out)
			client.conn.Close()
			delete(users, event.address)

		case message:
			for address, client := range users {
				if address == event.address {
					continue
				}
				client.out <- Outgoing{isFile: false, message: event.message}
			}
		case file:
			for address, client := range users {
				if address == event.address {
					continue
				}
				client.out <- Outgoing{isFile: true, file: event.file}
			}
		}
	}

}

func handleConnection(conn net.Conn, events chan<- Event) {

	defer conn.Close()

	reader := bufio.NewReader(conn)

	address := conn.RemoteAddr().String()

	e := Event{
		kind:    message,
		address: address,
		conn:    conn}

	bodyMessage := "nuevo usuario conectado"

	fmt.Printf("%s -> %s\n", bodyMessage, e.address)

	headers := make(map[string]string)

	headers["content-type"] = "text/plain"
	headers["size"] = strconv.Itoa(len(bodyMessage))
	headers["sender"] = "servidor"

	payload := buildHeader(headers)

	e.message = payload + bodyMessage
	events <- e

	for {
		headers, header, err := readHeaders(reader)

		if err != nil {
			if err == io.EOF {
				println("Usuario desconectado: " + address)
				e.kind = disconnection
				events <- e
				return
			}
			fmt.Println(err)
			return
		}

		switch headers["content-type"] {
		case "text/plain":
			size, err := strconv.Atoi(headers["size"])

			if err != nil {
				fmt.Println(err)
				continue
			}

			buffer := make([]byte, size)

			_, err = io.ReadFull(reader, buffer)

			if err != nil {
				fmt.Println(err)
				continue
			}

			body := string(buffer)
			header.WriteString(body)

			e.kind = message
			e.message = header.String()
			e.user = headers["sender"]

			events <- e

			fmt.Printf("%s-%s: %s\n", e.address, headers["sender"], body)
		case "file":

			f, err := handleFile(reader, headers)

			if err != nil {
				fmt.Println(err)
				continue
			}

			e.kind = file
			e.user = headers["sender"]
			e.file = f
			events <- e
		}

		e = Event{address: address, conn: conn}

	}

}
