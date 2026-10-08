package client

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/terminal"
)

var Ip string
var User string
var outputMu sync.Mutex

func printOutput(format string, args ...any) {
	outputMu.Lock()
	defer outputMu.Unlock()
	fmt.Printf(format, args...)
}

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

	in := make(chan string)
	out := make(chan string)

	go handleWrite(conn, in)
	go handleRead(conn, out)

	terminal.Run(in, out)
}

func handleRead(conn net.Conn, out chan<- string) {

	reader := bufio.NewReader(conn)

	for {
		headers, _, err := protocolchat.ReadHeaders(reader)

		if err != nil {
			out <- fmt.Sprintf("%s\n", err.Error())
			return
		}

		size, ok := headers["size"]

		if !ok {
			out <- "header size no existe.\n"
			return
		}

		n, err := strconv.Atoi(size)

		if err != nil {
			out <- fmt.Sprintf("Error en cast de header size: %s\n", err)
			return
		}

		buffer := make([]byte, n)

		kind, ok := headers["content-type"]

		if !ok {
			out <- "header content-type no existe.\n"
			return
		}

		switch kind {
		case "text/plain":

			n, err := reader.Read(buffer)

			if err != nil {
				out <- fmt.Sprintf("%s", err.Error())
				return
			}

			body := string(buffer[:n])

			out <- fmt.Sprintf("[sender: %s]\nmessage: %s\n", headers["sender"], body)

		case "file/notification":
			go downloadFile(headers)
		case "response":
			if msg, err := response(headers); err != nil {
				out <- fmt.Sprintf("%s\n", err.Error())
				return
			} else {
				out <- fmt.Sprintf("%s\n", *msg)
			}
		default:
			out <- "Error de cabezera 'content-type'= ¡no reconocido!\n"
		}
	}
}

func handleWrite(conn net.Conn, in <-chan string) {

	headers := make(map[string]string)

	for msg := range in {

		bodyMessage := msg

		switch bodyMessage {
		case "/file":

			err := sendFile(conn)

			if err != nil {
				printOutput("%s\n", err.Error())
				continue
			}

		default:
			headers["content-type"] = "text/plain"
			headers["size"] = strconv.Itoa(len(bodyMessage))
			headers["sender"] = User

			header := protocolchat.BuildHeader(headers)

			_, err := conn.Write([]byte(header + bodyMessage))

			if err != nil {
				printOutput("Error enviando mensaje: %s\n", err)
				return
			}

		}

	}
}
