package files

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lmmv9411/chago/internal/protocolfile"
)

func StartServer() {

	server, err := NewServer("tcp", ":8081")

	if err != nil {
		fmt.Println("Error al iniciar el servidor:", err)
		return
	}

	defer server.CloseConn()

	fmt.Println("Servidor archivos escuchando en el puerto 8081...")

	for {

		err := server.Accept()

		if err != nil {
			fmt.Println("Error al aceptar la conexión: ", err)
			continue
		}

		go handleConnection(server)
	}

}

func handleConnection(s *Server) {

	defer s.CloseConn()

	headers, err := protocolfile.ReadHeaders(s.Reader)

	if err != nil {
		msg := "Error en lecutra headers: " + err.Error()
		fmt.Println(msg)
		s.SendError(msg, RequestError)
		return
	}

	fileName, ok := headers["filename"]

	if !ok {
		msg := "falta header: 'filename'"
		fmt.Println(msg)
		s.SendError(msg, RequestError)
		return
	}

	sizeHeader, ok := headers["size"]

	if !ok {
		msg := "falta header: 'size'"
		fmt.Println(msg)
		s.SendError(msg, RequestError)
		return
	}

	method, ok := headers["method"]

	if !ok {
		msg := "falta header: 'method'"
		fmt.Println(msg)
		s.SendError(msg, RequestError)
		return
	}

	size, err := strconv.ParseInt(sizeHeader, 10, 64)

	if err != nil {
		msg := "Error en parsing de size"
		fmt.Println(msg)
		s.SendError(msg, RequestError)
		return
	}

	if size < 0 || size > protocolfile.MaxBodySize {
		msg := "Archivo con tamaño no permitido"
		fmt.Println(msg)
		s.SendError(msg, RequestError)

		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()
	if err != nil {
		msg := "Error al obtener directorio"
		fmt.Println(msg, err)
		s.SendError(msg, InternalError)
		return
	}

	err = os.MkdirAll(filepath.Join(currentDir, "uploads"), 0755)

	if err != nil {
		msg := "Error al crear directorio"
		fmt.Println(msg, err)
		s.SendError(msg, InternalError)
		return
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "uploads", safeFilename)

	switch method {
	case "upload":

		file, err := os.Create(filePath)

		if err != nil {
			msg := fmt.Sprintf("Error al crear archivo %s, %s\n", fileName, err.Error())
			fmt.Println(msg)
			s.SendError(msg, InternalError)
			return
		}

		defer file.Close()

		_, err = io.CopyN(file, s.Reader, size)

		if err != nil {
			msg := fmt.Sprintf("Error en stream %s, %s\n", fileName, err.Error())
			fmt.Println(msg)
			s.SendError(msg, InternalError)
			return
		}
	case "download":

		file, err := os.Open(filePath)

		if err != nil {
			msg := fmt.Sprintf("Error al leer archivo %s, %s\n", fileName, err.Error())
			fmt.Println(msg)
			s.SendError(msg, InternalError)
			return
		}

		defer file.Close()

		info, err := os.Stat(filePath)

		if err != nil {
			msg := fmt.Sprintln("Error al acceder a info de archivo: " + err.Error())
			fmt.Println(msg)
			s.SendError(msg, InternalError)
			return
		}

		_, err = io.CopyN(s.Conn, file, info.Size())

		if err != nil {
			msg := fmt.Sprintf("Error al crear archivo %s, %s\n", fileName, err.Error())
			fmt.Println(msg)
			s.SendError(msg, InternalError)
			return
		}
	default:
		msg := fmt.Sprintln("Método no existe: ", method)
		fmt.Println(msg)
		s.SendError(msg, InternalError)
	}

}
