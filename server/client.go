package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

func StartClient() {

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Escribir direccion ip ó [ y ] para usar por defecto: ")

	ip, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("Error al leer ip", err)
		return
	}

	ip = strings.TrimSpace(ip)

	if ip != "y" {
		if net.ParseIP(ip) == nil {
			fmt.Println("¡Dirección ip inválida!...")
			return
		}
	} else {
		ip = "192.168.1.33"
	}

	fmt.Print("Escribir usuario: ")
	user, err := reader.ReadString('\n')

	user = strings.TrimSpace(user)

	if err != nil {
		fmt.Println("Error al leer usuario", err)
		return
	}

	conn, err := net.Dial("tcp", ip+":8080")

	if err != nil {
		fmt.Println("Error al conectar:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Conectado al servidor")

	go handleWrite(conn, user)

	r := bufio.NewReader(conn)

	for {
		headers, _, err := readHeaders(r)

		if err != nil {
			fmt.Println(err)
			return
		}

		n, err := strconv.Atoi(headers["size"])

		if err != nil {
			fmt.Println(err)
			continue
		}

		kind := headers["content-type"]

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
			filename := headers["filename"]
			size, err := strconv.ParseInt(headers["size"], 10, 64)
			if err != nil {
				println("Error parsing header size", err)
				continue
			}
			sender := headers["sender"]

			fmt.Println(sender + " envió archivo: " + filename)

			err = os.MkdirAll("./downloads", 0755)

			if err != nil {
				println("Error al crear directorio downloads")
				continue
			}

			file, err := os.Create("./downloads/" + filename)

			if err != nil {
				println("Error al crear archivo: " + filename)
				continue
			}

			defer file.Close()

			_, err = io.CopyN(file, r, size)

			if err != nil {
				println("Error al escribir bytes del stream en archivo: " + filename)
				continue
			}

			fmt.Println("Archivo recibido: " + filename)
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
			err := sendFile(conn, user, scanner)
			if err != nil {
				fmt.Println(err)
			}
		default:
			headers["content-type"] = "text/plain"
			headers["size"] = strconv.Itoa(len(bodyMessage))
			headers["sender"] = user

			header := buildHeader(headers)

			_, err := conn.Write([]byte(header + bodyMessage))

			if err != nil {
				fmt.Println("Error enviando mensaje:", err)
				return
			}
		}

	}
}
