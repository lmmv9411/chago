package server

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type eventType int

const (
	connection eventType = iota
	disconnection
	message
)

type event struct {
	kind    eventType
	address string
	conn    net.Conn
	message string
}

type user struct {
	address string
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

	buffer := make([]byte, 1024)

	e := event{
		kind:    message,
		address: conn.RemoteAddr().String(),
		conn:    conn}

	msg := "nuevo usuario conectado"

	fmt.Printf("%s -> %s\n", msg, e.address)

	msgJson := Message{
		Sender:  "server",
		Content: fmt.Sprintf("%s -> %s\n", msg, e.address),
		Type:    TypeText,
		Size:    int64(len(msg)),
	}

	txt, err := msgJson.ToJson()

	if err != nil {
		fmt.Println("Error serializar mensaje")
		return
	}

	e.message = string(append(txt, '\n'))

	events <- e

	for {
		n, err := conn.Read(buffer)

		if err != nil {

			msg = fmt.Sprintf("Error: %s\n", err)

			if errors.Is(err, io.EOF) {
				msg = fmt.Sprintf("usuario %s cerró la conexión.\n", e.address)
				e.kind = disconnection
				events <- e
			}

			msgJson = Message{
				Sender:  "server",
				Content: msg,
				Type:    TypeText,
				Size:    int64(len(msg)),
			}

			txt, err := msgJson.ToJson()

			if err != nil {
				fmt.Println("Error serializar mensaje")
				return
			}

			e.message = string(append(txt, '\n'))
			e.kind = message
			events <- e
			fmt.Print(msg)
			break
		}

		e.message = string(buffer[:n])

		events <- e

		fmt.Printf("%s: %s\n", e.address, e.message)

	}
}
