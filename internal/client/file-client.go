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
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmmv9411/chago/internal/protocolchat"
	"github.com/lmmv9411/chago/internal/protocolfile"
	"github.com/lmmv9411/chago/internal/serverfiles"
	"github.com/lmmv9411/chago/internal/terminal"
)

func sendToServer(filePath string, connMsg net.Conn, events chan<- tea.Msg) {

	info, err := os.Stat(filePath)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al acceder al archivo: %v", err))
		return
	}

	if info.IsDir() {
		events <- terminal.ErrorChatMsg("Es un directorio")
		return
	}

	file, err := os.Open(filePath)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al abrir el archivo: %v", err))
		return
	}

	defer file.Close()

	headers := make(map[string]string)

	headers["filename"] = info.Name()
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["method"] = "upload"

	headersStr := protocolfile.BuildHeader(headers)

	conn, err := net.Dial("tcp", IP+":8081")

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al conectar a servidor files: %v", err))
		return
	}

	defer conn.Close()

	_, err = conn.Write([]byte(headersStr))

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al enviar header al servidor-files: %v", err))
		return
	}

	r := bufio.NewReader(conn)

	if _, err := isOk(r); err != nil {
		events <- terminal.ErrorChatMsg(err.Error())
		return
	}

	progress := &ProgressWriter{
		writer:   conn,
		total:    info.Size(),
		id:       "upload:" + filePath,
		filename: info.Name(),
		events:   events,
	}

	_, err = io.CopyN(progress, file, info.Size())

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al enviar archivo al servidor-files: %v", err))
		return
	}

	if _, err := isOk(r); err != nil {
		events <- terminal.ErrorChatMsg(err.Error())
		return
	}

	events <- terminal.NewTransferDoneMsg(progress.id, progress.filename)

	headers = make(map[string]string)
	headers["content-type"] = "file/notification"
	headers["size"] = strconv.FormatInt(info.Size(), 10)
	headers["sender"] = User
	headers["filename"] = info.Name()

	header := protocolchat.BuildHeader(headers)

	_, err = connMsg.Write([]byte(header))

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error enviando file/notification: %v", err))
		return
	}

}

func downloadFile(headers map[string]string, events chan<- tea.Msg) {

	sizeS := headers["size"]

	fileName, ok := headers["filename"]

	if !ok {
		events <- terminal.ErrorChatMsg("sin header filename")
		return
	}

	sender, ok := headers["sender"]

	if !ok {
		events <- terminal.ErrorChatMsg("sin header sender")
		return
	}

	size, err := strconv.ParseInt(sizeS, 10, 64)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error cast header size: %v", err))
		return
	}

	if size < 0 || size > protocolfile.GiB {
		events <- terminal.ErrorChatMsg("Archivo excede tamaño permitido")
		return
	}

	//Por El momento en el directorio donde se ejecuta luego se centralizaria
	currentDir, err := os.Getwd()

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al obtener directorio: %v", err))
		return
	}

	err = os.MkdirAll(filepath.Join(currentDir, "downloads"), 0755)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al crear directorio: %v", err))
		return
	}

	safeFilename := filepath.Base(fileName)
	filePath := filepath.Join(currentDir, "downloads", safeFilename)

	file, err := os.Create(filePath)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al crear archivo: %v", err))
		return
	}

	defer file.Close()

	progress := &ProgressWriter{
		writer:   file,
		total:    size,
		id:       "download:" + sender + ":" + fileName,
		filename: fileName,
		events:   events,
	}
	events <- terminal.InconmingChatMsg(fmt.Sprintf("%s envio archivo: %s", sender, fileName))

	conn, err := net.Dial("tcp", IP+":8081")

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al conectar a servidor files: %v", err))
		return
	}

	defer conn.Close()

	headers["method"] = "download"

	header := protocolfile.BuildHeader(headers)

	//Request file from server-file
	_, err = conn.Write([]byte(header))

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error al enviar headers request: %v", err))
		return
	}

	_, err = io.CopyN(progress, conn, size)

	if err != nil {
		events <- terminal.ErrorChatMsg(fmt.Sprintf("Error en el stream de archivo: %v", err))
		return
	}

	r := bufio.NewReader(conn)

	if _, err := isOk(r); err != nil {
		events <- terminal.ErrorChatMsg(err.Error())
		return
	}

	events <- terminal.NewTransferDoneMsg(progress.id, progress.filename)
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

type ProgressWriter struct {
	total          int64
	writer         io.Writer
	written        int64
	id             string
	filename       string
	events         chan<- tea.Msg
	hasReported    bool
	lastProgressAt time.Time
}

func (p *ProgressWriter) Write(data []byte) (int, error) {
	n, err := p.writer.Write(data)
	if err != nil {
		return n, err
	}

	p.written += int64(n)
	if p.total > 0 {

		now := time.Now()
		ratio := float64(p.written) / float64(p.total)
		completed := p.written >= p.total

		if !p.hasReported || now.Sub(p.lastProgressAt) >= 150*time.Millisecond || completed {

			p.events <- terminal.NewTransferProgressMsg(
				p.id,
				p.filename,
				ratio,
			)
			p.lastProgressAt = now
			p.hasReported = true
		}
	}

	return n, nil
}
