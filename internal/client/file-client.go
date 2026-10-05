package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
	"github.com/lmmv9411/chago/internal/serverfiles"
)

func sendFile(scanner *bufio.Scanner, conn net.Conn) error {

	fmt.Print("Escribir ruta de archivo:")

	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return errors.New("Error al leer texto: " + err.Error())
	}

	filePath := scanner.Text()

	go sendToServer(filePath, conn)

	return nil
}

func sendToServer(filePath string, connMsg net.Conn) {

	info, err := os.Stat(filePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	if info.IsDir() {
		fmt.Println("Es un directorio")
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	headers := make(map[string]string)

	headers["filename"] = info.Name()
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["method"] = "upload"

	headersStr := protocolfile.BuildHeader(headers)

	conn, err := net.Dial("tcp", Ip+":8081")

	if err != nil {
		fmt.Printf("Error al conectar a servidor files: %v", err)
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte(headersStr))

	if err != nil {
		fmt.Println("Error al enviar header al servidor-files: %w", err)
		return
	}

	if err := isOk(conn); err != nil {
		fmt.Println(err)
		return
	}

	progress := &ProgressWriter{writer: conn, total: info.Size(), barWidth: 30}

	_, err = io.CopyN(progress, file, info.Size())

	if err != nil {
		fmt.Println("Error al enviar archivo al servidor-files: %w", err)
		return
	}

	if err := isOk(conn); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("\nArchivo enviado y recibido por el servidor.")

	headers = make(map[string]string)
	headers["content-type"] = "file/notification"
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["sender"] = User
	headers["filename"] = info.Name()

	header := protocolchat.BuildHeader(headers)

	_, err = connMsg.Write([]byte(header))

	if err != nil {
		fmt.Println("Error enviando file/notification: ", err.Error())
		return
	}

}

func downloadFile(headers map[string]string) {

	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		fmt.Println("sin header filename")
		return
	}

	sender, ok := headers["filename"]

	if !ok {
		fmt.Println("sin header sender")
		return
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)

	if err != nil {
		fmt.Println("Error cast header size")
		return
	}

	if size < 0 || size > protocolfile.GiB {
		fmt.Println("Archivo excede tamaño permitido")
		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		fmt.Println("Error al obtener directorio")
		return
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0755)

	if err != nil {
		fmt.Println("Error al crear directorio")
		return
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.Create(filePath)

	if err != nil {
		fmt.Println("Error al crear archivo.")
		return
	}

	defer file.Close()

	progress := &ProgressWriter{total: size, writer: file, barWidth: 30}

	fmt.Printf("Recibiendo archivo %s de %s\n", fileName, sender)

	conn, err := net.Dial("tcp", Ip+":8081")

	if err != nil {
		fmt.Printf("Error al conectar a servidor files: %v", err)
		return
	}

	defer conn.Close()

	_, err = io.CopyN(progress, conn, size)

	if err != nil {
		fmt.Println("Error en el stream de archivo")
		return
	}

	fmt.Printf("Archivo recibido de %s: %s", sender, fileName)
}

func isOk(conn net.Conn) error {
	r := bufio.NewReader(conn)

	headers, err := protocolfile.ReadHeaders(r)

	if err != nil {
		return fmt.Errorf("Error al leer respuesta del servidor: %w", err)
	}

	status, ok := headers["status"]

	if !ok {
		return errors.New("respuesta del servidor sin status")
	}

	code, err := strconv.Atoi(status)

	if err != nil {
		return fmt.Errorf("status inválido en respuesta del servidor: %w", err)
	}

	if serverfiles.Status(code) != serverfiles.Ok {
		return fmt.Errorf("servidor respondió %s: %s", status, headers["message"])
	}

	return nil
}

type ProgressWriter struct {
	total    int64
	writer   io.Writer
	written  int64
	barWidth int
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err
	}

	p.written += int64(n)
	percentage := float64(p.written) / float64(p.total) * 100
	filled := int((percentage / 100.0) * float64(p.barWidth))

	fmt.Print("\r[")

	for i := range p.barWidth {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}

	fmt.Printf("] %.0f%%", percentage)

	return n, nil
}
