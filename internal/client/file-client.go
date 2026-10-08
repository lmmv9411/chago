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

func sendFile(conn net.Conn) error {

	scanner := bufio.NewScanner(os.Stdin)

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
		printOutput("%s\n", err.Error())
		return
	}

	if info.IsDir() {
		printOutput("Es un directorio\n")
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		printOutput("%s\n", err.Error())
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
		printOutput("Error al conectar a servidor files: %v\n", err)
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte(headersStr))

	if err != nil {
		printOutput("Error al enviar header al servidor-files: %s\n", err)
		return
	}

	r := bufio.NewReader(conn)

	if _, err := isOk(r); err != nil {
		printOutput("%s\n", err)
		return
	}

	progress := &ProgressWriter{writer: conn, total: info.Size(), barWidth: 30}

	_, err = io.CopyN(progress, file, info.Size())

	if err != nil {
		printOutput("Error al enviar archivo al servidor-files: %s\n", err)
		return
	}

	if _, err := isOk(r); err != nil {
		printOutput("%s\n", err)
		return
	}

	printOutput("\nArchivo enviado y recibido por el servidor.\n")

	headers = make(map[string]string)
	headers["content-type"] = "file/notification"
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["sender"] = User
	headers["filename"] = info.Name()

	header := protocolchat.BuildHeader(headers)

	_, err = connMsg.Write([]byte(header))

	if err != nil {
		printOutput("Error enviando file/notification: %s\n", err.Error())
		return
	}

}

func downloadFile(headers map[string]string) {

	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		printOutput("sin header filename\n")
		return
	}

	sender, ok := headers["sender"]

	if !ok {
		printOutput("sin header sender\n")
		return
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)

	if err != nil {
		printOutput("Error cast header size\n")
		return
	}

	if size < 0 || size > protocolfile.GiB {
		printOutput("Archivo excede tamaño permitido\n")
		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		printOutput("Error al obtener directorio\n")
		return
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0755)

	if err != nil {
		printOutput("Error al crear directorio\n")
		return
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.Create(filePath)

	if err != nil {
		printOutput("Error al crear archivo.\n")
		return
	}

	defer file.Close()

	progress := &ProgressWriter{total: size, writer: file, barWidth: 30}

	printOutput("%s envio archivo: %s\n", sender, fileName)

	conn, err := net.Dial("tcp", Ip+":8081")

	if err != nil {
		printOutput("Error al conectar a servidor files: %v\n", err)
		return
	}

	defer conn.Close()

	headers["method"] = "download"

	header := protocolfile.BuildHeader(headers)

	//Request file from server-file
	_, err = conn.Write([]byte(header))

	if err != nil {
		printOutput("Error al enviar headers request\n")
		return
	}

	_, err = io.CopyN(progress, conn, size)

	if err != nil {
		printOutput("Error en el stream de archivo\n")
		return
	}

	r := bufio.NewReader(conn)

	if _, err := isOk(r); err != nil {
		printOutput("%v\n", err)
		return
	}

	printOutput("\nArchivo recibido de %s: %s\n", sender, fileName)
}

func isOk(r *bufio.Reader) (*string, error) {

	headers, err := protocolfile.ReadHeaders(r)

	if err != nil {
		return nil, fmt.Errorf("Error al leer respuesta del servidor: %v\n", err)
	}

	status, ok := headers["status"]

	if !ok {
		return nil, errors.New("respuesta del servidor sin status")
	}

	code, err := strconv.Atoi(status)

	if err != nil {
		return nil, fmt.Errorf("status inválido en respuesta del servidor: %v\n", err)
	}

	if serverfiles.Status(code) != serverfiles.Ok {
		return nil, fmt.Errorf("servidor respondió %s: %s", status, headers["message"])
	}

	msg := headers["message"]
	return &msg, nil
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

	printOutput("] %.0f%%", percentage)

	return n, nil
}
