package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/lmmv9411/chago/internal/events"
	"github.com/lmmv9411/chago/internal/protocolfile"
)

func (c *Client) download(headers map[string]string) error {

	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		return fmt.Errorf("sin header filename")
	}

	sender, ok := headers["sender"]

	if !ok {
		return fmt.Errorf("sin header sender")
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)

	if err != nil {
		return fmt.Errorf("Error cast header size: %w", err)
	}

	if size < 0 || size > protocolfile.GiB {
		return fmt.Errorf("Archivo excede tamaño permitido")
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		return fmt.Errorf("Error al obtener directorio: %w", err)
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0755)

	if err != nil {
		return fmt.Errorf("Error al crear directorio: %w", err)
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.Create(filePath)

	if err != nil {
		return fmt.Errorf("Error al crear archivo: %w", err)
	}

	defer file.Close()

	progress := &ProgressWriter{
		writer:   file,
		total:    size,
		id:       "download:" + sender + ":" + fileName,
		filename: fileName,
		events:   c.events,
	}
	c.events <- events.Event{
		Kind:    events.ChatMessage,
		Message: fmt.Sprintf("%s envio archivo: %s", sender, fileName),
	}

	server, err := NewConnection(c.conn.address, ":8081")

	if err != nil {
		return fmt.Errorf("Error al conectar a servidor files: %w", err)
	}

	defer server.conn.Close()

	headers["method"] = "download"

	header := protocolfile.BuildHeader(headers)

	_, err = server.conn.Write([]byte(header))

	if err != nil {
		return fmt.Errorf("Error al enviar headers request: %w", err)
	}

	_, err = io.CopyN(progress, server.conn, size)

	if err != nil {
		return fmt.Errorf("Error en el stream de archivo: %w", err)
	}

	if _, err := isOk(server.reader); err != nil {
		return err
	}

	c.events <- events.Event{
		Kind:       events.TransferDone,
		TransferId: progress.id,
		FileName:   progress.filename,
	}
	return nil
}
