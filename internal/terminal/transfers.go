package terminal

import (
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

func NewTransferProgressMsg(id, filename string, ratio float64) tea.Msg {
	return TransferProgressMsg{
		id:       id,
		filename: filename,
		ratio:    ratio,
	}
}

func NewTransferDoneMsg(id, filename string) tea.Msg {
	return TransferDoneMsg{id: id, filename: filename}
}
