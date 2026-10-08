package terminal

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

type TransferInfo struct {
	id       string
	filename string
	percent  float64
	progress progress.Model
}

type TransferProgressMsg struct {
	id       string
	filename string
	ratio    float64
}

type TransferDoneMsg struct {
	id       string
	filename string
}

type MultiProgressReader struct {
	id       string
	filename string
	reader   io.Reader
	total    int64
	current  int64
	events   chan<- tea.Msg
}

func (pr *MultiProgressReader) Read(p []byte) (int, error) {

	n, err := pr.reader.Read(p)

	pr.current += int64(n)

	if pr.total > 0 {
		pr.events <- TransferProgressMsg{
			id:       pr.id,
			filename: pr.filename,
			ratio:    float64(pr.current) / float64(pr.total),
		}
	}

	return n, err
}

func download(events chan<- tea.Msg) tea.Cmd {

	return func() tea.Msg {

		size := int64(1024 * 100) //100KiB

		url := fmt.Sprintf("https://httpbin.org/bytes/%d", size)
		filepath := "archivo_descargado.bin"

		resp, err := http.Get(url)
		if err != nil {
			return ErrorChatMsg(fmt.Sprintf("Error al realizar la petición: %v", err))
		}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ErrorChatMsg(fmt.Sprintf("Servidor retornó estado: %s", resp.Status))
		}

		dest, err := os.Create(filepath)
		if err != nil {
			return ErrorChatMsg(fmt.Sprintf("Error al crear el archivo local: %v", err))
		}
		defer dest.Close()

		reader := &MultiProgressReader{
			id:       "archivo.bin",
			filename: "archivo.bin",
			total:    size,
			current:  0,
			reader:   resp.Body,
		}

		_, err = io.CopyN(dest, reader, size)

		if err != nil {
			return ErrorChatMsg(
				fmt.Sprintf(
					"Error durante el streaming del archivo: %v",
					err,
				),
			)
		}

		events <- TransferDoneMsg{id: reader.id, filename: reader.filename}

		return nil
	}
}
