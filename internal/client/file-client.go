package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/lmmv9411/chago/internal/files"
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

	if !isOk(conn) {
		return
	}

	if err != nil {
		fmt.Println("Error al enviar header al servidor-files: " + err.Error())
		return
	}

	progress := &ProgressWriter{writer: conn}

	_, err = io.CopyN(progress, file, size)

	if !isOk(conn) {
		return
	}

	if err != nil {
		fmt.Println("Error al enviar archivo al servidor-files: " + err.Error())
		return
	}

	fmt.Println("Archivo enviado.")
}

func isOk(conn net.Conn) bool {

	r := bufio.NewReader(conn)

	headers, _, err := protocolchat.ReadHeaders(r)

	if err != nil {
		fmt.Println(err)
	}

	code, err := strconv.Atoi(headers["status"])

	if err != nil {
		fmt.Println(err)
	}

	if files.Status(code) != files.Ok {
		fmt.Printf("%s: %s\n", headers["status"], headers["message"])
		return false
	}

	return true
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
