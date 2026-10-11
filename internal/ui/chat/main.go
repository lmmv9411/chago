package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
)

type (
	IncomingChatMsg string
	ErrorChatMsg    string
)

type chatmodel struct {
	transfers      map[string]*TransferInfo
	transfersOrder []string
	messages       []string

	textinput textinput.Model
	viewport  viewport.Model

	outgoing chan<- string
	events   chan tea.Msg
}

func initChatModel(outgoing chan<- string, events chan tea.Msg) chatmodel {
	return initChatStyle(outgoing, events)
}

func waitForEvent(events <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-events

		if !ok {
			return nil
		}
		return msg
	}
}

func (m chatmodel) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		waitForEvent(m.events),
	)
}

func (m chatmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		m, c := m.handleKey(msg, &cmds)
		if m != nil {
			return m, c
		}
	case IncomingChatMsg:
		m.appendMessage(string(msg))
	case TransferProgressMsg:
		m.handleTransferProgress(msg, &cmds)
	case TransferDoneMsg:
		m.handleTransferDone(msg)
	case ErrorChatMsg:
		m.appendMessage(errorChatStyle.Render(fmt.Sprintf("[Sistema]: %s", msg)))
	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
	}

	switch msg.(type) {
	case ErrorChatMsg, TransferDoneMsg,
		IncomingChatMsg, TransferProgressMsg:
		cmds = append(cmds, waitForEvent(m.events))
	}

	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	for _, t := range m.transfers {
		pm, cmd := t.progress.Update(msg)
		t.progress = pm.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *chatmodel) refreshViewport() {

	lines := strings.Join(m.messages, "\n\n")

	lines = ansi.Wrap(lines, max(m.viewport.Width-2, 1), "")
	m.viewport.SetContent(lines)
}

func (m *chatmodel) appendMessage(msg string) {
	if strings.HasPrefix(msg, "[yo]:") {
		msg = messageStyle.Render(msg)
	} else {
		msg = messageOtherStyle.Render(msg)
	}
	m.messages = append(m.messages, msg)
	m.refreshViewport()
	m.viewport.GotoBottom()
}

func (m chatmodel) View() string {
	return m.handleView()
}

func sendChatMessage(outgoing chan<- string, message string) tea.Cmd {
	return func() tea.Msg {
		outgoing <- message
		return nil
	}
}

func Run(outdoing chan<- string, events chan tea.Msg) {
	p := tea.NewProgram(initChatModel(outdoing, events), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error al ejecutar la aplicación: %v", err)
	}
}
