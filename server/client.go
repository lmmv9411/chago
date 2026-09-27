package server

import (
	"bufio"
	"fmt"
	"net"
	"os"
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

	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {

		byteMessages := scanner.Bytes()

		msg, err := FromJson(byteMessages)

		if err != nil {
			fmt.Println("Error al deserializar json.", err)
			continue
		}

		fmt.Printf("user   : %s\nmessage: %s\n\n", msg.Sender, msg.Content)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error de lectura en la conexión:", err)
	} else {
		fmt.Println("El host remoto cerró la conexión (EOF).")
	}

}

func handleWrite(conn net.Conn, user string) {

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {

		if err := scanner.Err(); err != nil {
			fmt.Println("Error al leer texto: ", err)
			continue
		}

		mensaje := scanner.Text()

		msg := Message{
			Sender:  user,
			Content: mensaje,
			Type:    TypeText,
			Size:    int64(len(mensaje)),
		}

		bytesJson, err := msg.ToJson()

		if err != nil {
			fmt.Println("Error en serialización de json.")
			break
		}

		_, err = conn.Write(append(bytesJson, '\n'))

		if err != nil {
			fmt.Println("Error enviando:", err)
			break
		}

	}
}
