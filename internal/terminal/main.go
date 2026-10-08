package terminal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inconmingChatMsg string

type model struct {
	transfers     map[string]*TransferInfo
	messages      []string
	textarea      textarea.Model
	viewport      viewport.Model
	outgoing      chan<- string
	incoming      <-chan string
	titleStyle    lipgloss.Style
	chatBoxStyle  lipgloss.Style
	inputBoxStyle lipgloss.Style
	errorStyle    lipgloss.Style
}

func initialModel(outgoing chan<- string, incoming <-chan string) model {
	ta := textarea.New()
	ta.Placeholder = "Escribir mensaje y presionar Enter..."
	ta.Focus()
	ta.SetHeight(1)
	ta.SetWidth(60)
	ta.ShowLineNumbers = false

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
		Padding(0, 1).
		Width(60)

	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	return model{
		messages:      []string{"[Sistema]: ¡Bienvenido al Chat!"},
		textarea:      ta,
		viewport:      vp,
		outgoing:      outgoing,
		incoming:      incoming,
		chatBoxStyle:  chatBoxStyle,
		titleStyle:    titleStyle,
		inputBoxStyle: inputBoxStyle,
		errorStyle:    errorStyle,
		transfers:     make(map[string]*TransferInfo),
	}

}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
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
			v := strings.TrimSpace(m.textarea.Value())

			if v == "" {
				return m, nil
			}

			if strings.HasPrefix(v, "/download") {
				m.textarea.Reset()
				return m, download()
			}

			m.messages = append(m.messages, fmt.Sprintf("[yo]: %s", v))
			m.textarea.Reset()

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			return m, sendChatMessage(m.outgoing, v)
		case tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	case inconmingChatMsg:
		m.messages = append(m.messages, string(msg))

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	case TransferProgressMsg:
		t, exist := m.transfers[msg.id]
		if !exist {
			prog := progress.New(
				progress.WithDefaultGradient(),
				progress.WithWidth(60),
			)
			t = &TransferInfo{id: msg.id, filename: msg.filename, progress: prog}
			m.transfers[msg.id] = t
		}
		t.percent = msg.ratio
		cmd = t.progress.SetPercent(t.percent)
		cmds = append(cmds, cmd)
	case TransferDoneMsg:
		if t, existe := m.transfers[msg.id]; existe {
			m.messages = append(m.messages, fmt.Sprintf("[Sistema]: %s finalizada.", t.filename))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			delete(m.transfers, msg.id)
		}
	case transferErrorMsg:
		m.messages = append(m.messages, m.errorStyle.Render(fmt.Sprintf("[Sistema]: %s", msg.err)))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	case tea.WindowSizeMsg:

		chatHeight := max(msg.Height-(11), 1)
		chatWidth := max(msg.Width-4, 1)

		m.viewport.Width = chatWidth
		m.viewport.Height = chatHeight

		m.chatBoxStyle = m.chatBoxStyle.Width(chatWidth)
		m.titleStyle = m.titleStyle.Width(msg.Width)
		m.textarea.SetWidth(chatWidth - 4)
		m.inputBoxStyle = m.inputBoxStyle.Width(chatWidth)

	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	for _, t := range m.transfers {
		pm, cmd := t.progress.Update(msg)
		t.progress = pm.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	viewSections := []string{
		m.titleStyle.Render("--- CHAGO ---"),
		m.chatBoxStyle.Render(m.viewport.View()),
	}

	if len(m.transfers) > 0 {
		for _, t := range m.transfers {
			linea := fmt.Sprintf(
				"[%s]: %s\n",
				t.filename,
				t.progress.ViewAs(t.percent),
			)
			viewSections = append(viewSections, linea)
		}
	}

	viewSections = append(
		viewSections,
		m.inputBoxStyle.Render(m.textarea.View()),
		"(Presiona Ctrl+C o Esc para salir)",
	)

	return strings.Join(viewSections, "\n")
}

var p *tea.Program

func WriteMsg(msg string) tea.Cmd {
	p.Send(inconmingChatMsg(msg))
	return nil
}

func sendChatMessage(outgoing chan<- string, message string) tea.Cmd {
	return func() tea.Msg {
		outgoing <- message
		return nil
	}
}

func Run(inc chan<- string, outgoing <-chan string) {

	p = tea.NewProgram(initialModel(inc, outgoing), tea.WithAltScreen())

	go incomingMessages(outgoing)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error al ejecutar la aplicación: %v", err)
	}

}

func incomingMessages(incoming <-chan string) {

	for msg := range incoming {
		p.Send(inconmingChatMsg(msg))
	}

}
