package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	btnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")). // Gris oscuro
			Background(lipgloss.Color("252")). // Gris claro
			Padding(0, 2)

	btnActiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")). // Blanco brillante
			Background(lipgloss.Color("63")).  // Azul llamativo
			Bold(true).
			Padding(0, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")).
			Width(40).
			AlignHorizontal(lipgloss.Center)

	errorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
)

const (
	focusAddress = iota
	focusUser
	focusOk
	focusCancel
)

type loginmodel struct {
	address    textinput.Model
	user       textinput.Model
	focusIndex int
	done       bool
	result     LoginData
	errorMsg   string
}

type LoginData struct {
	Address   string
	User      string
	cancelled bool
}

func initLoginModel() loginmodel {
	addres := textinput.New()
	addres.Placeholder = "192.168.1.33"
	addres.Prompt = "Dirección Ip > "
	addres.Cursor.SetMode(cursor.CursorBlink)
	addres.Focus()
	addres.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	addres.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Color del placeholder (Gris)
	addres.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))         // Color del texto escrito
	addres.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	user := textinput.New()
	user.Placeholder = "leo"
	user.Prompt = "     Usuario > "
	user.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	user.PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Color del placeholder (Gris)
	user.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))         // Color del texto escrito
	user.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return loginmodel{
		user:       user,
		address:    addres,
		focusIndex: focusAddress,
	}
}

func (l loginmodel) Init() tea.Cmd {
	return nil
}

func (l loginmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.Type {

		case tea.KeyCtrlC, tea.KeyEsc:
			return l, tea.Quit

		case tea.KeyTab, tea.KeyShiftTab, tea.KeyDown, tea.KeyUp:
			msg := msg.Type

			if msg == tea.KeyUp || msg == tea.KeyShiftTab {
				l.focusIndex--
			} else {
				l.focusIndex++
			}

			if l.focusIndex > focusCancel {
				l.focusIndex = focusAddress
			} else if l.focusIndex < focusAddress {
				l.focusIndex = focusCancel
			}

			cmdAddress := l.updateFocus()

			return l, cmdAddress

		case tea.KeyEnter:

			switch l.focusIndex {

			case focusOk:

				if strings.TrimSpace(l.user.Value()) == "" {
					l.errorMsg = "⚠️ El nombre de usuario no puede estar vacío."
					l.focusIndex = focusUser
					return l, l.updateFocus()
				}

				data := LoginData{
					Address:   strings.TrimSpace(l.address.Value()),
					User:      strings.TrimSpace(l.user.Value()),
					cancelled: false,
				}

				l.done = true
				l.result = data

				return l, tea.Quit
			case focusCancel:
				l.done = true
				return l, tea.Quit
			}
		default:
			l.errorMsg = ""
		}
	}

	// Enviar eventos de teclado al input que tiene el foco actual
	var cmd tea.Cmd

	switch l.focusIndex {
	case focusAddress:
		l.address, cmd = l.address.Update(msg)
	case focusUser:
		l.user, cmd = l.user.Update(msg)
	}

	return l, cmd

}

func (l *loginmodel) updateFocus() tea.Cmd {
	if l.focusIndex == focusAddress {
		l.address.Focus()
	} else {
		l.address.Blur()
	}

	if l.focusIndex == focusUser {
		l.user.Focus()
	} else {
		l.user.Blur()
	}

	return textinput.Blink
}

func (l loginmodel) View() string {

	warningSection := ""
	if l.errorMsg != "" {
		warningSection = errorStyle.Render(l.errorMsg) + "\n"
	}

	viewSections := []string{
		titleStyle.Render("=== Login ===\n"),
		l.address.View(),
		l.user.View() + " " + warningSection + "\n",
	}

	btnOk := btnStyle.Render(" OK ")
	btnCancel := btnStyle.Render(" Cancel ")

	switch l.focusIndex {
	case focusOk:
		btnOk = btnActiveStyle.Render(" OK ")
	case focusCancel:
		btnCancel = btnActiveStyle.Render(" Cancel ")
	}

	viewSections = append(viewSections,
		btnOk+" "+btnCancel+"\n\n",
		lipgloss.NewStyle().
			Foreground(
				lipgloss.Color("241")).
			Render("Usar Tab para navegar, Enter para Seleccionar\n",
				"Ctrl + C ó ESC para salir."),
	)

	return strings.Join(viewSections, "\n")
}

func LoginMain() *LoginData {

	p := tea.NewProgram(initLoginModel(), tea.WithAltScreen())

	finalModel, err := p.Run()

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
		return nil
	}

	l, ok := finalModel.(loginmodel)
	if !ok || !l.done {
		fmt.Println("El formulario se cerró inesperadamente.")
		return nil
	}

	return &l.result

}
