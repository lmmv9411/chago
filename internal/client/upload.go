package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
)

func (c *Client) Upload(filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("Error al acceder al archivo: %w", err)
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("la ruta no es un archivo regular")
	}
	if strings.ContainsAny(info.Name(), "\r\n") {
		return fmt.Errorf("el nombre del archivo contiene caracteres no permitidos")
	}
	if info.Size() > protocolfile.MaxBodySize {
		return fmt.Errorf("el archivo excede el límite de %d bytes", protocolfile.MaxBodySize)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Error al abrir el archivo: %w", err)
	}

	defer file.Close()

	headers := make(map[string]string)

	headers["filename"] = info.Name()
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["method"] = "upload"

	headersStr := protocolfile.BuildHeader(headers)

	server, err := NewConnection(c.conn.address, "8081")
	if err != nil {
		return fmt.Errorf("Error al conectar a servidor files: %w", err)
	}

	defer server.conn.Close()
	stopWatching := closeConnectionOnDone(c.done, server.conn)
	defer stopWatching()

	_, err = server.Write([]byte(headersStr))
	if err != nil {
		return fmt.Errorf("Error al enviar header al servidor-files: %w", err)
	}

	if _, err := isOk(server.reader); err != nil {
		return err
	}

	progress := &ProgressWriter{
		writer:   server.conn,
		total:    info.Size(),
		id:       c.newTransferID("upload"),
		filename: info.Name(),
		emit:     c.emit,
	}

	_, err = io.CopyN(progress, file, info.Size())
	if err != nil {
		return fmt.Errorf("Error al enviar archivo al servidor-files: %w", err)
	}

	if _, err := isOk(server.reader); err != nil {
		return err
	}

	headers = make(map[string]string)
	headers["content-type"] = "file/notification"
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["sender"] = c.name
	headers["filename"] = info.Name()

	header := protocolchat.BuildHeader(headers)

	_, err = c.conn.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("Error enviando file/notification: %w", err)
	}

	c.emit(chat.Event{
		Kind:       chat.TransferDone,
		TransferId: progress.id,
		FileName:   progress.filename,
	})

	return nil
}

func isOk(r *bufio.Reader) (*string, error) {
	headers, err := protocolfile.ReadHeaders(r)
	if err != nil {
		return nil, fmt.Errorf("error al leer respuesta del servidor: %v\n", err)
	}

	status, ok := headers["status"]

	if !ok {
		return nil, errors.New("respuesta del servidor sin status")
	}

	code, err := strconv.Atoi(status)
	if err != nil {
		return nil, fmt.Errorf("status inválido en respuesta del servidor: %v\n", err)
	}

	if protocolfile.Status(code) != protocolfile.Ok {
		return nil, fmt.Errorf("servidor respondió %s: %s", status, headers["message"])
	}

	msg := headers["message"]
	return &msg, nil
}
