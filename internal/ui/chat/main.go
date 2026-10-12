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
	"github.com/lmmv9411/chago/internal/api"
	"github.com/lmmv9411/chago/internal/events"
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

	s      api.Api
	events chan events.Event
}

func waitForEvent(ev chan events.Event) tea.Cmd {
	return func() tea.Msg {
		event, ok := <-ev

		if !ok {
			return nil
		}

		switch event.Kind {
		case events.ChatMessage:
			return IncomingChatMsg(event.Message)
		case events.Error:
			return ErrorChatMsg(event.Message)
		case events.TransferDone:
			return TransferDoneMsg{
				id:       event.TransferId,
				filename: event.FileName,
			}
		case events.TransferProgress:
			return TransferProgressMsg{
				id:       event.TransferId,
				filename: event.FileName,
				ratio:    event.Ratio,
			}
		}

		return event
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

func Run(s api.Api) error {

	p := tea.NewProgram(
		initChatStyle(s, s.Events()),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("Error al ejecutar la aplicación: %v", err)
	}

	return nil
}
