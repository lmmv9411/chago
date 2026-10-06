package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/protocolchat"
)

var Ip string
var User string
var terminal = &Terminal{}

func StartClient() {

	r := bufio.NewReader(os.Stdin)

	fmt.Print("Escribir direccion ip ó [ y ] para usar por defecto: ")

	inputIP, err := r.ReadString('\n')

	if err != nil {
		fmt.Println("Error al leer ip", err)
		return
	}

	inputIP = strings.TrimSpace(inputIP)

	if inputIP == "y" {
		Ip = "192.168.1.33"
	} else {
		if net.ParseIP(inputIP) == nil {
			fmt.Println("¡Dirección ip inválida!...")
			return
		}
		Ip = inputIP

	}

	fmt.Print("Escribir nombre de usuario: ")
	inputUser, err := r.ReadString('\n')

	if err != nil {
		fmt.Println("Error al leer usuario", err)
		return
	}

	User = strings.TrimSpace(inputUser)

	conn, err := net.Dial("tcp", Ip+":8080")

	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Conectado al servidor")

	go handleWrite(conn)

	reader := bufio.NewReader(conn)

	for {
		headers, _, err := protocolchat.ReadHeaders(reader)

		if err != nil {
			terminal.Print("%s\n", err.Error())
			return
		}

		size, ok := headers["size"]

		if !ok {
			terminal.Print("header size no existe.\n")
			return
		}

		n, err := strconv.Atoi(size)

		if err != nil {
			terminal.Print("Error en cast de header size: %s\n", err)
			return
		}

		buffer := make([]byte, n)

		kind, ok := headers["content-type"]

		if !ok {
			terminal.Print("header content-type no existe.\n")
			return
		}

		switch kind {
		case "text/plain":

			n, err := reader.Read(buffer)

			if err != nil {
				terminal.Print("%s", err.Error())
				return
			}

			body := string(buffer[:n])

			terminal.Print("---------------------------------------\n")
			terminal.Print("sender: %s\nmessage: \n", headers["sender"])
			terminal.Print("%s\n", body)
			terminal.Print("---------------------------------------\n")

		case "file/notification":
			go downloadFile(headers)
		case "response":
			if msg, err := response(headers); err != nil {
				terminal.Print("%s\n", err.Error())
				return
			} else {
				terminal.Print("%s\n", *msg)
			}
		default:
			terminal.Print("Error de cabezera 'content-type'= ¡no reconocido!\n")
		}
	}

}

func handleWrite(conn net.Conn) {

	scanner := bufio.NewScanner(os.Stdin)
	headers := make(map[string]string)

	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			terminal.Print("Error al leer texto: %s\n", err)
			continue
		}

		bodyMessage := scanner.Text()

		switch bodyMessage {
		case "/file":

			err := sendFile(scanner, conn)

			if err != nil {
				terminal.Print("%s\n", err.Error())
				continue
			}

		default:
			headers["content-type"] = "text/plain"
			headers["size"] = strconv.Itoa(len(bodyMessage))
			headers["sender"] = User

			header := protocolchat.BuildHeader(headers)

			_, err := conn.Write([]byte(header + bodyMessage))

			if err != nil {
				terminal.Print("Error enviando mensaje: %s\n", err)
				return
			}

		}

	}
}
