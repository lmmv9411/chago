package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

func (m *chatmodel) handleWindowSize(msg tea.WindowSizeMsg) {

	chatHeight := max(msg.Height-7, 1)
	chatWidth := max(msg.Width-4, 1)

	m.viewport.Width = chatWidth
	m.viewport.Height = chatHeight

	chatBoxStyle = chatBoxStyle.Width(chatWidth)
	titleChatStyle = titleChatStyle.Width(msg.Width)
	m.textinput.Width = chatWidth - 4
	inputBoxStyle = inputBoxStyle.Width(chatWidth)

	wasAtBottom := m.viewport.AtBottom()

	m.refreshViewport()

	if wasAtBottom {
		m.viewport.GotoBottom()
	}

	if len(m.transfers) > 0 {
		for _, t := range m.transfers {
			t.progress.Width = max(m.viewport.Width-2, 1)
		}
	}
}

func (m *chatmodel) handleTransferDone(msg TransferDoneMsg) {

	if t, existe := m.transfers[msg.id]; existe {

		m.appendMessage(fmt.Sprintf("[Sistema]: %s finalizada.", t.filename))

		delete(m.transfers, msg.id)

		for i, id := range m.transfersOrder {
			if id == msg.id {
				m.transfersOrder = append(m.transfersOrder[:i], m.transfersOrder[i+1:]...)
			}
		}
	}
}

func (m *chatmodel) handleKey(msg tea.KeyMsg, cmds *[]tea.Cmd) (tea.Model, tea.Cmd) {

	var cmd tea.Cmd

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyEnter:
		v := strings.TrimSpace(m.textinput.Value())

		if v == "" {
			return m, nil
		}

		m.appendMessage(fmt.Sprintf("[yo]: %s", v))
		m.textinput.Reset()

		if after, ok := strings.CutPrefix(v, "/file"); ok {
			filePath := strings.TrimSpace(after)
			if filePath == "" {
				*cmds = append(*cmds, func() tea.Msg {
					return ErrorChatMsg("Escribir la ruta del archivo")
				})
			} else {
				*cmds = append(*cmds, runClientOperation(func() error {
					return m.s.Upload(filePath)
				}))
			}
		} else {
			*cmds = append(*cmds, runClientOperation(func() error {
				return m.s.Send(v)
			}))
		}
	case tea.KeyPgUp, tea.KeyPgDown:
		m.viewport, cmd = m.viewport.Update(msg)
		*cmds = append(*cmds, cmd)
	}

	return nil, nil
}

func runClientOperation(operation func() error) tea.Cmd {
	return func() tea.Msg {
		if err := operation(); err != nil {
			return ErrorChatMsg(err.Error())
		}
		return nil
	}
}

func (m *chatmodel) handleTransferProgress(msg TransferProgressMsg) {
	t, exist := m.transfers[msg.id]
	if !exist {
		prog := progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(max(m.viewport.Width-2, 1)),
		)
		t = &TransferInfo{id: msg.id, filename: msg.filename, progress: prog}
		m.transfers[msg.id] = t
		m.transfersOrder = append(m.transfersOrder, msg.id)
	}
	t.percent = msg.ratio
}
