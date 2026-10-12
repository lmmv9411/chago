package chat

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/lmmv9411/chago/internal/api"
)

var (
	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("63")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	messageOtherStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("36")).
				Background(lipgloss.Color("235")).
				Padding(0, 1)

	titleChatStyle = lipgloss.NewStyle().
			Bold(true).
			Width(60).
			Foreground(lipgloss.Color("86")).
			Align(lipgloss.Center)

	chatBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1).
			Width(60)

	inputBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(0, 1)

	errorChatStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	progressStyle = lipgloss.NewStyle().Padding(0, 1)
)

func initModel(s api.Api) chatmodel {
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

	return chatmodel{
		messages:  []string{"[Sistema]: ¡Bienvenido al Chat!"},
		transfers: make(map[string]*TransferInfo),

		textinput: ti,
		viewport:  vp,
		s:         s,
	}
}

func (m chatmodel) handleView() string {
	viewSections := []string{
		titleChatStyle.Render("--- CHAGO ---"),
		chatBoxStyle.Render(m.viewport.View()),
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
				progressStyle.Render(t.progress.ViewAs(t.percent)),
			)
			viewSections = append(viewSections, linea)
		}
	}

	viewSections = append(
		viewSections,
		inputBoxStyle.Render(m.textinput.View()),
		"(Presiona Ctrl+C o Esc para salir)",
	)

	return strings.Join(viewSections, "\n")
}
