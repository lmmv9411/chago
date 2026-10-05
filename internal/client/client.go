package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/protocolchat"
)

var Ip string

func StartClient() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Escribir direccion ip ó [ y ] para usar por defecto: ")

	inputIP, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error al leer ip", err)
		return
	}

	inputIP = strings.TrimSpace(inputIP)

	if inputIP == "y" {
		Ip = "192.168.1.33"
	} else {
		if net.ParseIP(Ip) == nil {
			fmt.Println("¡Dirección ip inválida!...")
			return
		}
		Ip = inputIP

	}

	fmt.Print("Escribir usuario: ")
	user, err := reader.ReadString('\n')

	user = strings.TrimSpace(user)

	if err != nil {
		fmt.Println("Error al leer usuario", err)
		return
	}

	conn, err := net.Dial("tcp", Ip+":8080")

	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Conectado al servidor")

	go handleWrite(conn, user)

	r := bufio.NewReader(conn)

	for {
		headers, _, err := protocolchat.ReadHeaders(r)

		if err != nil {
			fmt.Println(err)
			return
		}

		size, ok := headers["size"]

		if !ok {
			fmt.Println("header size no existe.")
			return
		}

		n, err := strconv.Atoi(size)

		if err != nil {
			fmt.Println("Error en cast de header size: ", err)
			return
		}

		kind, ok := headers["content-type"]

		if !ok {
			fmt.Println("header content-type no existe.")
			return
		}

		switch kind {
		case "text/plain":
			fmt.Println("---------------------------------------")
			fmt.Printf("sender: %s\nmessage: ", headers["sender"])
			_, err := io.CopyN(os.Stdout, r, int64(n))
			fmt.Println()
			fmt.Println("---------------------------------------")
			if err != nil {
				fmt.Println(err)
			}
		case "file":
			err = downloadFile(headers, r)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	}

}

func handleWrite(conn net.Conn, user string) {

	scanner := bufio.NewScanner(os.Stdin)
	headers := make(map[string]string)

	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			fmt.Println("Error al leer texto: ", err)
			continue
		}

		bodyMessage := scanner.Text()

		switch bodyMessage {
		case "/file":

			filename, err := sendFile(scanner)

			if err != nil {
				fmt.Println(err)
				continue
			}

			headers["content-type"] = "file/notification"
			headers["size"] = strconv.Itoa(len(bodyMessage))
			headers["sender"] = user
			headers["filename"] = filename

			header := protocolchat.BuildHeader(headers)

			_, err = conn.Write([]byte(header))

			if err != nil {
				fmt.Println("Error enviando mensaje: ", err)
				return
			}

		default:
			headers["content-type"] = "text/plain"
			headers["size"] = strconv.Itoa(len(bodyMessage))
			headers["sender"] = user

			header := protocolchat.BuildHeader(headers)

			_, err := conn.Write([]byte(header + bodyMessage))

			if err != nil {
				fmt.Println("Error enviando mensaje: ", err)
				return
			}
		}

	}
}
