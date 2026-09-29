package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
)

type eventType int

const (
	connection eventType = iota
	disconnection
	message
	file
)

type event struct {
	kind    eventType
	address string
	conn    net.Conn
	message string
}

func StartServer() {

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Servidor escuchando en el puerto 8080...")

	events := make(chan event, 5)
	go server(events)

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión:", err)
			continue
		}

		e := event{
			kind:    connection,
			conn:    conn,
			address: conn.RemoteAddr().String()}

		events <- e

		go handleConnection(conn, events)
	}
}

func server(events <-chan event) {

	users := make(map[string]net.Conn)

	for event := range events {

		switch event.kind {
		case connection:
			users[event.address] = event.conn

		case disconnection:
			event.conn.Close()
			delete(users, event.address)

		case message:

			for direccion, conn := range users {

				if direccion != event.address {
					_, err := io.WriteString(conn, event.message)
					if err != nil {
						fmt.Printf("Error al enviar mensaje a %s\n", direccion)
					}
				}
			}
		}
	}

}

func handleConnection(conn net.Conn, events chan<- event) {

	defer conn.Close()

	reader := bufio.NewReader(conn)

	e := event{
		kind:    message,
		address: conn.RemoteAddr().String(),
		conn:    conn}

	msg := "nuevo usuario conectado"

	fmt.Printf("%s -> %s\n", msg, e.address)

	payload := "content-type:text/plain\n" +
		"size:" + strconv.Itoa(len(msg)) + "\n" +
		"sender:servidor" + "\n\n" +
		msg

	e.message = payload
	events <- e

	for {
		var header strings.Builder
		headers := make(map[string]string)
		for {
			line, err := reader.ReadString('\n')
			header.WriteString(line)

			if err != nil {
				fmt.Println(err)
				events <- event{
					kind:    disconnection,
					address: conn.RemoteAddr().String(),
					conn:    conn,
				}
				return
			}

			if line == "\n" {
				break
			}

			data := strings.SplitN(line, ":", 2)
			if len(data) != 2 {
				fmt.Println("Error en protocolo de cabecera, no contiene par llave - valor.", data)
				events <- event{
					kind:    disconnection,
					address: conn.RemoteAddr().String(),
					conn:    conn,
				}
				return
			}
			headers[strings.TrimSpace(data[0])] = strings.TrimSpace(data[1])
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

		}

	}

}
