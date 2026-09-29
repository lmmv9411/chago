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

func handleFile(reader *bufio.Reader, headers map[string]string) (bool, error) {
	filename := headers["filename"]
	size, err := strconv.ParseInt(headers["size"], 10, 64)
	if err != nil {
		return false, errors.New("Error al convertir size to int: " + headers["size"] + err.Error())
	}

	err = os.MkdirAll("./uploads", 0755)

	if err != nil {
		fmt.Printf("error al crear el directorio: %v\n", err)
		return false, err
	}

	file, err := os.Create("./uploads/" + filename)
	if err != nil {
		return false, errors.New("Error al crear fichero: " + filename + err.Error())
	}

	_, err = io.CopyN(file, reader, size)
	if err != nil {
		return false, errors.New("Error al copiar fichero: " + filename + err.Error())
	}

	return true, nil
}

func sendFile(writer io.Writer, user string) {
	fmt.Print("Escribir ruta: ")
	r := bufio.NewReader(os.Stdin)
	filePath, err := r.ReadString('\n')
	if err != nil {
		println("Error:", err)
		return
	}
	filePath = strings.TrimSpace(filePath)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("el archivo no existe en la ruta: %s\n", filePath)
			return
		}
		fmt.Printf("error al acceder a la ruta: %v\n", err)
		return
	}
	if info.IsDir() {
		fmt.Println("la ruta especificada apunta a un directorio, no a un archivo")
		return
	}

	size := info.Size()

	headers := make(map[string]string)
	headers["content-type"] = "file"
	headers["size"] = strconv.FormatInt(size, 10)
	headers["sender"] = user
	headers["filename"] = info.Name()
	header := buildHeader(headers)

	_, err = writer.Write([]byte(header))
	if err != nil {
		fmt.Println("Error enviando header:", err)
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error abriendo archivo:", err)
		return
	}

	_, err = io.CopyN(writer, file, size)

	if err != nil {
		fmt.Println("Error enviando archivo:", err)
		return
	}

	fmt.Println("Archivo enviado con éxito")
}
