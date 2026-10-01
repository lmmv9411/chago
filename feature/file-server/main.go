package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/lmmv9411/chago/feature/file-server/server"
)

func main() {

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Escribir tipo: [ s ] server; [ c ] client: ")
	var text string

	if scanner.Scan() {
		text = scanner.Text()
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer la terminal. ", err)
		return
	}

	switch text {
	case "server", "s":
		server.StartServer()
	case "client", "c":
		server.StartClient()
	default:
		fmt.Println("¡Opción inválida!")
	}
}
