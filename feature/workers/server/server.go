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

			newClient := &Client{
				conn:    event.conn,
				out:     make(chan Outgoing, 10),
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

	bodyMessage := "nuevo usuario conectado"

	fmt.Printf("%s -> %s\n", bodyMessage, address)

	headers := make(map[string]string)

	headers["content-type"] = "text/plain"
	headers["size"] = strconv.Itoa(len(bodyMessage))
	headers["sender"] = "servidor"

	payload := buildHeader(headers)

	e := Event{
		kind:    message,
		address: address,
		message: payload + bodyMessage,
	}

	events <- e

	for {

		headers, header, err := readHeaders(reader)

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
			fmt.Println("Sin head 'Sender'")
			events <- Event{kind: disconnection, address: address}
			return
		}

		if len(sender) > maxSenderSize {
			fmt.Println("Valor head de 'sender' excede limite tamaño")
			events <- Event{kind: disconnection, address: address}
			return
		}

		switch headers["content-type"] {

		case "text/plain":
			size, err := strconv.Atoi(headers["size"])

			if err != nil {
				fmt.Println(err)
				continue
			}

			if size > maxBodySize {
				fmt.Println("Body mensaje exce Tamaño limite.")
				events <- Event{kind: disconnection, address: address}
				return
			}

			buffer := make([]byte, size)

			_, err = io.ReadFull(reader, buffer)

			if err != nil {
				fmt.Println(err)
				continue
			}

			body := string(buffer)
			header.WriteString(body)

			events <- Event{
				kind:    message,
				message: header.String(),
				user:    sender,
				address: address,
			}

			fmt.Printf("%s: %s\n", sender, body)

		case "file":

			fmt.Printf("recibiendo archivo de: %s", sender)

			f, err := handleFile(reader, headers)

			if err != nil {
				fmt.Println(err)
				continue
			}

			events <- Event{
				kind:    file,
				user:    sender,
				file:    f,
				address: address,
			}

			fmt.Printf("%s enviando archivo %s\n", sender, f.name)

		default:
			fmt.Println("Content-type desconocido: ", headers["content-type"])
			events <- Event{kind: disconnection, address: address}
			return
		}

	}

}
