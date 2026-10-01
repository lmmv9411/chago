package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func handleFile(reader *bufio.Reader, headers map[string]string) (FileEvent, error) {

	filename := headers["filename"]

	size, err := strconv.ParseInt(headers["size"], 10, 64)

	if err != nil {
		return FileEvent{}, errors.New("Error al convertir header size a int: " + headers["size"] + err.Error())
	}

	err = os.MkdirAll("./uploads", 0755)

	if err != nil {
		return FileEvent{}, errors.New("error al crear el directorio:" + err.Error())
	}

	file, err := os.Create("./uploads/" + filename)
	if err != nil {
		return FileEvent{}, errors.New("Error al crear fichero: " + filename + err.Error())
	}

	defer file.Close()

	_, err = io.CopyN(file, reader, size)
	if err != nil {
		return FileEvent{}, errors.New("Error al copiar fichero: " + filename + err.Error())
	}

	f := FileEvent{name: filename, size: size, path: "./uploads/" + filename}

	return f, nil
}

type ProgressWriter struct {
	total   int64
	writer  io.Writer
	written int64
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err
	}

	p.written += int64(n)
	percentage := float64(p.written) / float64(p.total) * 100
	barWidth := 30
	filled := int((percentage / 100.0) * float64(barWidth))

	fmt.Print("\r[")

	for i := range barWidth {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}

	fmt.Printf("] %.0f%%", percentage)

	return n, nil
}

func sendFile(writer io.Writer, user string, scanner *bufio.Scanner) error {

	fmt.Print("Escribir ruta: ")

	scanner.Scan()

	if err := scanner.Err(); err != nil {
		fmt.Println("Error al leer texto: ", err)
		return err
	}

	filePath := strings.TrimSpace(scanner.Text())

	info, err := os.Stat(filePath)

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("el archivo no existe en la ruta: %s\n", filePath)
			return err
		}
		fmt.Printf("error al acceder a la ruta: %v\n", err)
		return err
	}

	if info.IsDir() {
		return errors.New("la ruta especificada apunta a un directorio, no a un archivo")
	}

	size := info.Size()

	file, err := os.Open(filePath)

	if err != nil {
		fmt.Println("Error abriendo archivo:", err)
		return err
	}

	defer file.Close()

	headers := make(map[string]string)
	headers["content-type"] = "file"
	headers["size"] = strconv.FormatInt(size, 10)
	headers["sender"] = user
	headers["filename"] = info.Name()
	header := buildHeader(headers)

	_, err = writer.Write([]byte(header))

	if err != nil {
		fmt.Println("Error enviando header:", err)
		return err
	}

	fmt.Println("Enviando archivo")
	p := &ProgressWriter{
		writer: writer,
		total:  size,
	}

	_, err = io.CopyN(p, file, size)

	if err != nil {
		fmt.Println("Error enviando archivo:", err)
		return err
	}

	fmt.Println("Archivo enviado con éxito")
	return nil
}

func handleFileClient(headers map[string]string, r io.Reader) error {
	filename := headers["filename"]
	size, err := strconv.ParseInt(headers["size"], 10, 64)
	if err != nil {
		return errors.New("Error parsing header size. " + err.Error())
	}
	sender := headers["sender"]

	fmt.Println(sender + " envió archivo: " + filename)

	err = os.MkdirAll("./downloads", 0755)

	if err != nil {
		return errors.New("Error al crear directorio downloads")
	}

	file, err := os.Create("./downloads/" + filename)

	if err != nil {
		return errors.New("Error al crear archivo: " + filename)
	}

	defer file.Close()

	p := &ProgressWriter{
		writer: file,
		total:  size,
	}

	_, err = io.CopyN(p, r, size)

	if err != nil {
		return errors.New("Error al escribir bytes del stream en archivo: " + filename)
	}

	fmt.Println("Archivo recibido: " + filename)

	return nil
}
