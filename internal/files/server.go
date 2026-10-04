package files

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
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

	reader := bufio.NewReader(conn)

	headers, err := protocolfile.ReadHeaders(reader)

	if err != nil {
		fmt.Println("Error en lecutra headers: ", err)
		return
	}

	fileName := headers["filename"]
	sizeHeader := headers["size"]
	method := headers["method"]

	size, err := strconv.ParseInt(sizeHeader, 10, 64)
	if err != nil {
		fmt.Println("Error en parsing de size")
		return
	}

	if size > protocolfile.MaxBodySize {
		fmt.Println("Archivo supera tamaño maximo permitido")
		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error al obtener directorio: ", err)
		return
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "uploads", safeFilename)

	switch method {
	case "upload":

		file, err := os.Create(filePath)

		if err != nil {
			fmt.Printf("Error al crear archivo %s, %s\n", fileName, err.Error())
			return
		}

		_, err = io.CopyN(file, reader, size)

		if err != nil {
			fmt.Printf("Error al crear archivo %s, %s\n", fileName, err.Error())
			return
		}
	case "download":

		file, err := os.Open(filePath)

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
