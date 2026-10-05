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

	"github.com/lmmv9411/chago/internal/protocolfile"
	"github.com/lmmv9411/chago/internal/serverfiles"
)

func sendFile(scanner *bufio.Scanner) (string, error) {

	fmt.Print("Escribir ruta de archivo:")
	scanner.Scan()

	if err := scanner.Err(); err != nil {
		return "", errors.New("Error al leer texto: " + err.Error())
	}

	filePath := scanner.Text()

	file, err := os.Open(filePath)

	if err != nil {
		return "", err
	}
	defer file.Close()

	info, err := os.Stat(filePath)

	if err != nil {
		return "", err
	}

	if info.IsDir() {
		return "", errors.New("Es un directorio")
	}

	headers := make(map[string]string)

	headers["filename"] = info.Name()
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["method"] = "upload"

	headersStr := protocolfile.BuildHeader(headers)

	return info.Name(), sendToServer(headersStr, file, info.Size())
}

func downloadFile(headers map[string]string, reader *bufio.Reader) error {

	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		return errors.New("sin header filename")
	}

	sender, ok := headers["filename"]

	if !ok {
		return errors.New("sin header sender")
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)

	if err != nil {
		return errors.New("Error cast header size")
	}

	if size < 0 || size > protocolfile.GiB {
		return errors.New("Archivo excede tamaño permitido")
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		return errors.New("Error al obtener directorio")
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0755)

	if err != nil {
		return errors.New("Error al crear directorio")
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.Create(filePath)

	if err != nil {
		return errors.New("Error al crear archivo.")
	}

	defer file.Close()

	progress := &ProgressWriter{total: size, writer: file, barWidth: 30}

	fmt.Printf("Recibiendo archivo %s de %s", fileName, sender)

	_, err = io.CopyN(progress, reader, int64(size))

	if err != nil {
		return errors.New("Error en el stream de archivo")
	}

	fmt.Printf("Archivo recibido de %s: %s", sender, fileName)

	return nil
}

func sendToServer(header string, file *os.File, size int64) error {
	println("Este es la ip: ", Ip)
	conn, err := net.Dial("tcp", Ip+":8081")

	if err != nil {
		return fmt.Errorf("Error al conectar a servidor files: %w", err)
	}

	defer conn.Close()

	_, err = conn.Write([]byte(header))

	if err != nil {
		return fmt.Errorf("Error al enviar header al servidor-files: %w", err)
	}

	if err := isOk(conn); err != nil {
		fmt.Println(err)
		return err
	}

	progress := &ProgressWriter{writer: conn, total: size, barWidth: 30}

	_, err = io.CopyN(progress, file, size)

	if err != nil {
		return fmt.Errorf("Error al enviar archivo al servidor-files: %w", err)
	}

	if err := isOk(conn); err != nil {
		fmt.Println(err)
		return err
	}

	fmt.Println("\nArchivo enviado y recibido por el servidor.")
	return nil
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
