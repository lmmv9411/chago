package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/lmmv9411/chago/ejercicios"
	"github.com/lmmv9411/chago/server"
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
	case "read":
		ejercicios.StartRead()
	case "clousure":
		ejercicios.InitClousure()
	case "stream":
		ejercicios.Stream()
	default:
		fmt.Println("¡Opción inválida!")
	}
}
