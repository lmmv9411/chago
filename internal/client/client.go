package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/terminal"
)

var Ip string
var User string

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
		if net.ParseIP(inputIP) == nil {
			fmt.Println("¡Dirección ip inválida!...")
			return
		}
		Ip = inputIP

	}

	fmt.Print("Escribir nombre de usuario: ")
	inputUser, err := reader.ReadString('\n')

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

	incoming := make(chan string)
	events := make(chan tea.Msg)

	go handleWrite(conn, incoming, events)
	go handleRead(conn, events)

	terminal.Run(incoming, events)
}

func handleRead(conn net.Conn, events chan tea.Msg) {

	reader := bufio.NewReader(conn)

	for {
		headers, _, err := protocolchat.ReadHeaders(reader)

		if err != nil {
			events <- terminal.ErrorChatMsg(fmt.Sprintf("%s\n", err.Error()))
			return
		}

		size, ok := headers["size"]

		if !ok {
			events <- terminal.ErrorChatMsg("header size no existe.\n")
			return
		}

		n, err := strconv.Atoi(size)

		if err != nil {
			events <- terminal.ErrorChatMsg(fmt.Sprintf("Error en cast de header size: %s\n", err))
			return
		}

		buffer := make([]byte, n)

		kind, ok := headers["content-type"]

		if !ok {
			events <- terminal.ErrorChatMsg("header content-type no existe.\n")
			return
		}

		switch kind {
		case "text/plain":

			n, err := reader.Read(buffer)

			if err != nil {
				events <- terminal.ErrorChatMsg(fmt.Sprintf("%s", err.Error()))
				return
			}

			body := string(buffer[:n])

			events <- terminal.InconmingChatMsg(fmt.Sprintf("[%s]: %s", headers["sender"], body))

		case "file/notification":
			go downloadFile(headers, events)
		case "response":
			if _, err := response(headers); err != nil {
				events <- terminal.ErrorChatMsg(fmt.Sprintf("%s\n", err.Error()))
				return
				// } else {
				// 	events <- terminal.InconmingChatMsg(fmt.Sprintf("%s\n", *msg))
			}
		default:
			events <- terminal.ErrorChatMsg("Error de cabezera 'content-type'= ¡no reconocido!\n")
		}
	}
}

func handleWrite(conn net.Conn, incoming chan string, events chan tea.Msg) {

	headers := make(map[string]string)

	for msg := range incoming {

		bodyMessage := msg

		if after, ok := strings.CutPrefix(bodyMessage, "/file"); ok {

			filePath := strings.TrimSpace(after)

			if filePath == "" {
				events <- terminal.ErrorChatMsg("Escribir la ruta del archivo")
				continue
			}

			go sendToServer(filePath, conn, events)

			continue
		}

		headers["content-type"] = "text/plain"
		headers["size"] = strconv.Itoa(len(bodyMessage))
		headers["sender"] = User

		header := protocolchat.BuildHeader(headers)

		_, err := conn.Write([]byte(header + bodyMessage))

		if err != nil {
			events <- terminal.ErrorChatMsg(fmt.Sprintf("Error enviando mensaje: %s\n", err))
			return
		}

	}
}
