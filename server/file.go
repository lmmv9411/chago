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
		fmt.Println("la ruta especificada apunta a un directorio, no a un archivo")
		return err
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

	_, err = io.CopyN(writer, file, size)

	if err != nil {
		fmt.Println("Error enviando archivo:", err)
		return err
	}

	fmt.Println("Archivo enviado con éxito")
	return nil
}
