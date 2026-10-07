package terminal

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type inconmingChatMsg string
type idProgress string

type progressState struct {
	id       string
	filename string
	percent  float64
	progress progress.Model
	interval time.Duration
	step     float64
	//size int64
	//written int64
}

type model struct {
	messages      []string
	textarea      textarea.Model
	viewport      viewport.Model
	err           error
	bars          map[string]*progressState
	titleStyle    lipgloss.Style
	chatBoxStyle  lipgloss.Style
	inputBoxStyle lipgloss.Style
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		simulateDownload("xxx", m.bars["xxx"].interval),
		simulateDownload("yyy", m.bars["yyy"].interval),
		simulateIncomingChat(),
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

			if strings.HasPrefix(v, "/file") {
				// test
				return m, nil
			}

			m.messages = append(m.messages, fmt.Sprintf("[yo]: %s", v))
			m.textarea.Reset()

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			return m, nil
		case tea.KeyPgUp, tea.KeyPgDown:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	case idProgress:

		id := string(msg)

		prg, ok := m.bars[id]

		if !ok {
			return m, nil
		}

		prg.percent += prg.step

		if prg.percent >= 1.0 {
			prg.percent = 1.0
			m.messages = append(m.messages, "[Sistema]: ¡Descarga Completada!")
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		} else {
			cmds = append(cmds, simulateDownload(id, prg.interval))
		}

	case inconmingChatMsg:
		m.messages = append(m.messages, fmt.Sprintf("[Soporte]: %s", string(msg)))

		cmds = append(cmds, simulateIncomingChat())

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()
	case tea.WindowSizeMsg:

		activeBars := len(m.bars)

		chatHeight := max(msg.Height-(11+2*activeBars), 1)
		chatWidth := max(msg.Width-4, 1)

		m.viewport.Width = chatWidth
		m.viewport.Height = chatHeight

		m.chatBoxStyle = m.chatBoxStyle.Width(chatWidth)
		m.titleStyle = m.titleStyle.Width(msg.Width)
		m.textarea.SetWidth(chatWidth - 4)
		m.inputBoxStyle = m.inputBoxStyle.Width(chatWidth)

		m.messages = append(m.messages, fmt.Sprintf("[Soporte]: w%d h%d", msg.Width, msg.Height))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)

	for _, p := range m.bars {
		pm, cmd := p.progress.Update(msg)
		p.progress = pm.(progress.Model)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {

	viewSections := []string{
		m.titleStyle.Render("--- CHAGO ---"),
		m.chatBoxStyle.Render(m.viewport.View()),
	}

	for _, p := range m.bars {
		if p.percent < 1.0 {
			viewSections = append(viewSections, fmt.Sprintf(
				"Descargando %s: %s",
				p.filename,
				p.progress.ViewAs(p.percent)))
		} else {
			delete(m.bars, p.id)
		}
	}

	viewSections = append(
		viewSections,
		m.inputBoxStyle.Render(m.textarea.View()),
		"(Presiona Ctrl+C o Esc para salir)",
	)

	return strings.Join(viewSections, "\n")
}

func initialModel() model {
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

	prg := progressState{
		id:       "xxx",
		percent:  0,
		filename: "Christian bombole guitar.mp4",
		step:     0.05,
		interval: 600 * time.Millisecond,
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(30),
		)}

	prgii := progressState{
		id:       "yyy",
		percent:  0,
		filename: "Flamme Kapaya Solo Guitar.mp4",
		interval: 300 * time.Millisecond,
		step:     0.10,
		progress: progress.New(
			progress.WithDefaultGradient(),
			progress.WithWidth(30),
		)}

	bars := make(map[string]*progressState)
	bars["xxx"] = &prg
	bars["yyy"] = &prgii

	return model{
		messages:      []string{"[Sistema]: ¡Bienvenido al Chat!"},
		textarea:      ta,
		viewport:      vp,
		bars:          bars,
		chatBoxStyle:  chatBoxStyle,
		titleStyle:    titleStyle,
		inputBoxStyle: inputBoxStyle,
	}

}

func simulateDownload(id string, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return idProgress(id)
	})
}

func simulateIncomingChat() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return inconmingChatMsg("¡Hola tienes un nuevo mensaje de prueba!")
	})
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error al ejecutar la aplicación: %v", err)
	}
}
