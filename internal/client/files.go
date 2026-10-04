package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/lmmv9411/chago/internal/protocolchat"
)

func sendFile(scanner *bufio.Scanner) error {

	fmt.Print("Escribir ruta de archivo:")
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return errors.New("Error al leer texto: " + err.Error())
	}

	filePath := scanner.Text()

	file, err := os.Open(filePath)

	if err != nil {
		return err
	}

	info, err := os.Stat(filePath)

	if err != nil {
		return err
	}

	if info.IsDir() {
		return errors.New("Es un directorio")
	}

	headers := make(map[string]string)

	headers["filename"] = info.Name()
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["method"] = "upload"

	headersStr := protocolchat.BuildHeader(headers)

	go sendToServer(headersStr, file, info.Size())

	return nil
}

func sendToServer(header string, file *os.File, size int64) {
	defer file.Close()
	conn, err := net.Dial("tcp", Ip+":8081")
	if err != nil {
		fmt.Println("Error al conectar a servidor files: " + err.Error())
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(header))

	if err != nil {
		fmt.Println("Error al enviar header al servidor-files: " + err.Error())
		return
	}

	_, err = io.CopyN(conn, file, size)

	if err != nil {
		fmt.Println("Error al enviar archivo al servidor-files: " + err.Error())
		return
	}
}
