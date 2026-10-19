package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lmmv9411/chago/internal/chat"
	"github.com/lmmv9411/chago/internal/protocolfile"
)

func (c *Client) download(headers map[string]string) error {
	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		return fmt.Errorf("sin header filename")
	}

	if strings.ContainsAny(fileName, "\r\n") {
		return fmt.Errorf("nombre de archivo inválido")
	}

	sender, ok := headers["sender"]

	if !ok {
		return fmt.Errorf("sin header sender")
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)
	if err != nil {
		return fmt.Errorf("error cast header size: %w", err)
	}

	if size < 0 || size > protocolfile.GiB {
		return fmt.Errorf("archivo excede tamaño permitido")
	}

	// Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error al obtener directorio: %w", err)
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0o755)
	if err != nil {
		return fmt.Errorf("error al crear directorio: %w", err)
	}

	safeFilename := filepath.Base(strings.ReplaceAll(fileName, `\`, "/"))
	if safeFilename == "." || safeFilename == ".." || strings.TrimSpace(safeFilename) == "" {
		return fmt.Errorf("nombre de archivo inválido")
	}
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.CreateTemp(filepath.Dir(filePath), ".chago-download-*")
	if err != nil {
		return fmt.Errorf("Error al crear archivo: %w", err)
	}

	tempPath := file.Name()
	defer os.Remove(tempPath)
	defer file.Close()

	progress := &ProgressWriter{
		writer:   file,
		total:    size,
		id:       c.newTransferID("download"),
		filename: fileName,
		emit:     c.emit,
	}

	c.emit(chat.Event{
		Kind:    chat.ChatMessage,
		Message: fmt.Sprintf("[%s]: ↓ %s", sender, fileName),
	})

	server, err := NewConnection(c.conn.address, "8081")
	if err != nil {
		return fmt.Errorf("error al conectar a servidor files: %w", err)
	}

	defer func() {
		if err := server.conn.Close(); err != nil {
			c.emit(chat.Event{
				Message: fmt.Sprintf(
					"Error al cerrar conexion de servidor files %s",
					err),
				Kind: chat.Error,
			})
		}
	}()

	stopWatching := closeConnectionOnDone(c.done, server.conn)
	defer stopWatching()

	headers["method"] = "download"

	header := protocolfile.BuildHeader(headers)

	_, err = server.conn.Write([]byte(header))
	if err != nil {
		return fmt.Errorf("error al enviar headers request: %w", err)
	}

	_, err = io.CopyN(progress, server.conn, size)
	if err != nil {
		return fmt.Errorf("error en el stream de archivo: %w", err)
	}

	if _, err := isOk(server.reader); err != nil {
		return err
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("error al guardar archivo descargado: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("error al cerrar archivo descargado: %w", err)
	}
	if err := os.Rename(tempPath, filePath); err != nil {
		return fmt.Errorf("error al publicar archivo descargado: %w", err)
	}

	c.emit(chat.Event{
		Kind:       chat.TransferDone,
		TransferId: progress.id,
		FileName:   progress.filename,
	})
	return nil
}
