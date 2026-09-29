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

	Events := make(chan Event, 5)
	go server(Events)

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

		Events <- e

		go handleConnection(conn, Events)
	}
}

func server(events <-chan Event) {

	users := make(map[string]net.Conn)

	for Event := range events {

		switch Event.kind {
		case connection:
			users[Event.address] = Event.conn

		case disconnection:
			Event.conn.Close()
			delete(users, Event.address)

		case message:

			for direccion, conn := range users {

				if direccion != Event.address {
					_, err := io.WriteString(conn, Event.message)
					if err != nil {
						fmt.Printf("Error al enviar mensaje a %s\n", direccion)
					}
				}
			}
		}
	}

}

func handleConnection(conn net.Conn, events chan<- Event) {

	defer conn.Close()

	reader := bufio.NewReader(conn)

	e := Event{
		kind:    message,
		address: conn.RemoteAddr().String(),
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

			e.message = header.String()

			events <- e

			fmt.Printf("%s-%s: %s\n", e.address, headers["sender"], body)
		case "file":
			err := handleFile(reader, headers)
			if err != nil {
				fmt.Println(err)
				continue
			}
		}

	}

}
