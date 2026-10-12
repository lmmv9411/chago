package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
	"github.com/lmmv9411/chago/internal/serverfiles"
)

func (c *Client) Upload(filePath string) error {

	info, err := os.Stat(filePath)

	if err != nil {
		return fmt.Errorf("Error al acceder al archivo: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("Es un directorio")
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
		id:       "upload:" + filePath,
		filename: info.Name(),
		events:   c.events,
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

	c.events <- chat.Event{
		Kind:       chat.TransferDone,
		TransferId: progress.id,
		FileName:   progress.filename,
	}

	return nil
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
