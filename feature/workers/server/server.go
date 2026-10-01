package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
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

	users := make(map[string]*Client)

	for event := range events {

		switch event.kind {
		case connection:
			users[event.address] = &Client{
				conn: event.conn,
				out:  make(chan Outgoing),
			}
			go Worker(users[event.address])

		case disconnection:
			event.conn.Close()
			delete(users, event.address)

		case message:
			for address, client := range users {

				if address == event.address {
					continue
				}

				client.out <- Outgoing{isFile: false, message: event.message}

			}
		case file:
			file, err := os.Open(event.file.path)
			if err != nil {
				fmt.Println("Error abriendo fichero", err)
				continue
			}

			headers := make(map[string]string)
			headers["content-type"] = "file"
			headers["filename"] = event.file.name
			headers["size"] = strconv.FormatInt(event.file.size, 10)
			headers["sender"] = event.user
			header := buildHeader(headers)

			for address, client := range users {
				if address == event.address {
					continue
				}
				/*_, err := conn.Write([]byte(header))
				if err != nil {
					fmt.Println("Error escribiendo cabezeras de archivo.", err)
					continue
				}

				_, err = file.Seek(0, io.SeekStart)

				if err != nil {
					fmt.Println("Error reposicionando archivo:", err)
					continue
				}

				_, err = io.CopyN(conn, file, event.file.size)

				if err != nil {
					fmt.Println("Error escribiendo bytes de archivo en cliente", err)
					continue
				}*/

				client.out <- Outgoing{
					isFile: true,
					header: header,
					reader: file,
					size:   event.file.size}

			}

			file.Close()
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
