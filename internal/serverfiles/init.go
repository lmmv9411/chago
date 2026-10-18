package serverfiles

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/lmmv9411/chago/internal/protocolfile"
)

func StartServer() error {

	server, err := NewServer("tcp", ":8081")

	if err != nil {
		return fmt.Errorf("error al iniciar el servidor de archivos: %w", err)
	}

	defer server.CloseListener()

	fmt.Println("Servidor archivos escuchando en el puerto 8081...")

	var backoff time.Duration
	for {

		connection, err := server.Accept()

		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			if netErr, ok := err.(net.Error); ok && netErr.Temporary() {
				if backoff == 0 {
					backoff = 5 * time.Millisecond
				} else {
					backoff *= 2
				}
				if backoff > time.Second {
					backoff = time.Second
				}
				fmt.Println("Error temporal al aceptar conexión:", err)
				time.Sleep(backoff)
				continue
			}
			return fmt.Errorf("error al aceptar conexión de archivos: %w", err)
		}
		backoff = 0

		select {
		case activeConnections <- struct{}{}:
			go func() {
				defer func() { <-activeConnections }()
				handleConnection(connection)
			}()
		default:
			if err := connection.SendError("servidor ocupado", protocolfile.InternalError); err != nil {
				fmt.Println("Error rechazando conexión por capacidad:", err)
			}
			if err := connection.CloseConn(); err != nil {
				fmt.Println("Error cerrando conexión rechazada:", err)
			}
		}
	}
}

var activeConnections = make(chan struct{}, maxConcurrentConnections)

const maxConcurrentConnections = 16

func handleConnection(c *Connection) {

	defer c.CloseConn()

	headers, err := protocolfile.ReadHeaders(c.Reader)

	if err != nil {
		reject(c, "Error al leer headers: "+err.Error(), protocolfile.RequestError)
		return
	}

	fileName, ok := headers["filename"]

	if !ok {
		reject(c, "falta header: 'filename'", protocolfile.RequestError)
		return
	}
	if !validFilename(fileName) {
		reject(c, "nombre de archivo inválido", protocolfile.RequestError)
		return
	}

	sizeHeader, ok := headers["size"]

	if !ok {
		reject(c, "falta header: 'size'", protocolfile.RequestError)
		return
	}

	method, ok := headers["method"]

	if !ok {
		reject(c, "falta header: 'method'", protocolfile.RequestError)
		return
	}
	if method != "upload" && method != "download" {
		reject(c, "método no permitido", protocolfile.RequestError)
		return
	}

	size, err := strconv.ParseInt(sizeHeader, 10, 64)

	if err != nil {
		reject(c, "tamaño inválido", protocolfile.RequestError)
		return
	}

	if size < 0 || size > protocolfile.MaxBodySize {
		reject(c, "archivo con tamaño no permitido", protocolfile.RequestError)
		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		reject(c, "error al obtener directorio: "+err.Error(), protocolfile.InternalError)
		return
	}

	uploadDir := filepath.Join(currentDir, "uploads")
	err = os.MkdirAll(uploadDir, 0o755)

	if err != nil {
		reject(c, "error al crear directorio: "+err.Error(), protocolfile.InternalError)
		return
	}

	filePath := filepath.Join(uploadDir, fileName)

	switch method {
	case "upload":

		file, err := os.CreateTemp(uploadDir, ".chago-upload-*")

		if err != nil {
			reject(c, fmt.Sprintf("error al crear archivo: %s", err), protocolfile.InternalError)
			return
		}

		tempPath := file.Name()
		defer os.Remove(tempPath)

		if err := c.SendOk("ok"); err != nil {
			fmt.Println("Error al confirmar inicio de carga:", err)
			file.Close()
			return
		}

		_, err = io.CopyN(file, c.Reader, size)

		if err != nil {
			file.Close()
			reject(c, fmt.Sprintf("error al recibir archivo: %s", err), protocolfile.InternalError)
			return
		}

		if err := file.Sync(); err != nil {
			file.Close()
			reject(c, fmt.Sprintf("error al guardar archivo: %s", err), protocolfile.InternalError)
			return
		}
		if err := file.Close(); err != nil {
			reject(c, fmt.Sprintf("error al cerrar archivo: %s", err), protocolfile.InternalError)
			return
		}
		if err := os.Rename(tempPath, filePath); err != nil {
			reject(c, fmt.Sprintf("error al publicar archivo: %s", err), protocolfile.InternalError)
			return
		}
		if err := c.SendOk("Archivo recibido"); err != nil {
			fmt.Println("Error al confirmar archivo recibido:", err)
		}

	case "download":

		file, err := os.Open(filePath)

		if err != nil {
			reject(c, fmt.Sprintf("error al abrir archivo: %s", err), protocolfile.InternalError)
			return
		}

		defer file.Close()

		info, err := file.Stat()

		if err != nil {
			reject(c, fmt.Sprintf("error al consultar archivo: %s", err), protocolfile.InternalError)
			return
		}
		if !info.Mode().IsRegular() {
			reject(c, "el destino no es un archivo regular", protocolfile.RequestError)
			return
		}
		if info.Size() != size {
			reject(c, "el tamaño solicitado no coincide con el archivo", protocolfile.RequestError)
			return
		}

		_, err = io.CopyN(c.Conn, file, info.Size())

		if err != nil {
			fmt.Println("Error al enviar archivo:", err)
			return
		}

		if err := c.SendOk(fmt.Sprintf("Ok %s %d", fileName, info.Size())); err != nil {
			fmt.Println("Error al confirmar descarga:", err)
		}
	}

}

func validFilename(name string) bool {
	return name != "" &&
		name != "." &&
		name != ".." &&
		filepath.Base(name) == name &&
		!strings.ContainsAny(name, "/\\:") &&
		strings.IndexFunc(name, unicode.IsControl) < 0
}

func reject(c *Connection, message string, status protocolfile.Status) {
	fmt.Println(message)
	if err := c.SendError(message, status); err != nil {
		fmt.Println("Error al enviar respuesta de error:", err)
	}
}
