package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type (
	InconmingChatMsg string
	ErrorChatMsg     string
)

type model struct {
	transfers      map[string]*TransferInfo
	transfersOrder []string
	messages       []string

	textinput textinput.Model
	viewport  viewport.Model

	outgoing chan<- string
	events   chan tea.Msg

	titleStyle    lipgloss.Style
	chatBoxStyle  lipgloss.Style
	inputBoxStyle lipgloss.Style
	errorStyle    lipgloss.Style
	progressStyle lipgloss.Style
}

func initialModel(outgoing chan<- string, events chan tea.Msg) model {
	ti := textinput.New()
	ti.Placeholder = "Escribir mensaje y presionar Enter..."
	ti.Prompt = "> "
	ti.Cursor.SetMode(cursor.CursorBlink)
	ti.Focus()
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Color del placeholder (Gris)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))         // Color del texto escrito
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(60, 10)
	vp.SetContent("[Sistema]: ¡Bienvenido al Chat!")

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Width(60).
		Foreground(lipgloss.Color("86")).
		Align(lipgloss.Center)

	chatBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1).
		Width(60)

	inputBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	progressStyle := lipgloss.NewStyle().Padding(0, 1)

	return model{
		messages:  []string{"[Sistema]: ¡Bienvenido al Chat!"},
		transfers: make(map[string]*TransferInfo),

		textinput: ti,
		viewport:  vp,

		outgoing: outgoing,
		events:   events,

		chatBoxStyle:  chatBoxStyle,
		titleStyle:    titleStyle,
		inputBoxStyle: inputBoxStyle,
		errorStyle:    errorStyle,
		progressStyle: progressStyle,
	}
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

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		waitForEvent(m.events),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			v := strings.TrimSpace(m.textinput.Value())

			if v == "" {
				return m, nil
			}

			m.messages = append(m.messages, fmt.Sprintf("[yo]: %s", v))
			m.textinput.Reset()

			m.refreshViewport()
			m.viewport.GotoBottom()

			return m, sendChatMessage(m.outgoing, v)
		case tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	case InconmingChatMsg:
		m.messages = append(m.messages, string(msg))

		m.refreshViewport()
		m.viewport.GotoBottom()

	case TransferProgressMsg:
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
		cmd = t.progress.SetPercent(t.percent)
		cmds = append(cmds, cmd)
	case TransferDoneMsg:
		if t, existe := m.transfers[msg.id]; existe {
			m.messages = append(m.messages, fmt.Sprintf("[Sistema]: %s finalizada.", t.filename))
			m.refreshViewport()
			m.viewport.GotoBottom()

			delete(m.transfers, msg.id)

			for i, id := range m.transfersOrder {
				if id == msg.id {
					m.transfersOrder = append(m.transfersOrder[:i], m.transfersOrder[i+1:]...)
				}
			}
		}
	case ErrorChatMsg:
		m.messages = append(m.messages, m.errorStyle.Render(fmt.Sprintf("[Sistema]: %s", msg)))
		m.refreshViewport()
		m.viewport.GotoBottom()
	case tea.WindowSizeMsg:

		chatHeight := max(msg.Height-7, 1)
		chatWidth := max(msg.Width-4, 1)

		m.viewport.Width = chatWidth
		m.viewport.Height = chatHeight

		m.chatBoxStyle = m.chatBoxStyle.Width(chatWidth)
		m.titleStyle = m.titleStyle.Width(msg.Width)
		m.textinput.Width = chatWidth - 4
		m.inputBoxStyle = m.inputBoxStyle.Width(chatWidth)

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

	switch msg.(type) {
	case ErrorChatMsg, TransferDoneMsg, InconmingChatMsg, TransferProgressMsg:
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

func (m *model) refreshViewport() {
	content := strings.Join(m.messages, "\n")
	m.viewport.SetContent(ansi.Wrap(content, max(m.viewport.Width, 1), ""))
}

func (m model) View() string {
	viewSections := []string{
		m.titleStyle.Render("--- CHAGO ---"),
		m.chatBoxStyle.Render(m.viewport.View()),
	}

	if len(m.transfers) > 0 {
		for _, id := range m.transfersOrder {

			t, ok := m.transfers[id]
			if !ok {
				continue
			}

			linea := fmt.Sprintf(
				"[%s]\n%s\n",
				t.filename,
				m.progressStyle.Render(t.progress.ViewAs(t.percent)),
			)
			viewSections = append(viewSections, linea)
		}
	}

	viewSections = append(
		viewSections,
		m.inputBoxStyle.Render(m.textinput.View()),
		"(Presiona Ctrl+C o Esc para salir)",
	)

	return strings.Join(viewSections, "\n")
}

func sendChatMessage(outgoing chan<- string, message string) tea.Cmd {
	return func() tea.Msg {
		outgoing <- message
		return nil
	}
}

func Run(outdoing chan<- string, events chan tea.Msg) {
	p := tea.NewProgram(initialModel(outdoing, events), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error al ejecutar la aplicación: %v", err)
	}
}
