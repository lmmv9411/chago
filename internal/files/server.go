package files

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/lmmv9411/chago/internal/protocolfile"
)

func StartServer() {

	listener, err := net.Listen("tcp", ":8081")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer listener.Close()

	fmt.Println("Servidor archivos escuchando en el puerto 8081...")

	for {
		conn, err := listener.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión: ", err)
			continue
		}

		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {

	defer conn.Close()

	headers, err := protocolfile.ReadHeaders(conn)

	if err != nil {
		fmt.Println("Error en lecutra headers: ", err)
		return
	}

	fileName := headers["filename"]
	sizeHeader := headers["size"]
	event := headers["event"]

	size, err := strconv.ParseInt(sizeHeader, 10, 64)
	if err != nil {
		fmt.Println("Error en parsing de size")
		return
	}

	switch event {
	case "upload":

		file, err := os.Create("../storage/uploads/" + fileName)

		if err != nil {
			fmt.Printf("Error al crear archivo %s, %s\n", fileName, err.Error())
			return
		}

		_, err = io.CopyN(file, conn, size)

		if err != nil {
			fmt.Printf("Error al crear archivo %s, %s\n", fileName, err.Error())
			return
		}
	case "download":
		file, err := os.Open("../storage/uploads/" + fileName)
		if err != nil {
			fmt.Printf("Error al leer archivo %s, %s\n", fileName, err.Error())
			return
		}

		_, err = io.CopyN(conn, file, size)

		if err != nil {
			fmt.Printf("Error al crear archivo %s, %s\n", fileName, err.Error())
			return
		}

	}

}
