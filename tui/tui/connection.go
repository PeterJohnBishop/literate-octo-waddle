package tui

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "github.com/lib/pq"
)

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Render("[ Connect ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Connect"))
)

type Connection struct {
	focusIndex int
	inputs     []textinput.Model
	cursorMode cursor.Mode
	status     string // connection status message
	db         *sql.DB
}

type ConnectionSuccessMsg struct{}

func InitialConnectionModel() Connection {
	m := Connection{
		inputs: make([]textinput.Model, 5),
	}

	var t textinput.Model
	for i := range m.inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 32
		t.Width = 64

		// host, port, username, password, DB name
		switch i {
		case 0:
			t.Placeholder = "host"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
		case 1:
			t.Placeholder = "port"
			t.CharLimit = 64
		case 2:
			t.Placeholder = "username"
			t.CharLimit = 64
		case 3:
			t.Placeholder = "password"
			t.EchoMode = textinput.EchoPassword
			t.EchoCharacter = '•'
		case 4:
			t.Placeholder = "db name"
			t.CharLimit = 64
		}

		m.inputs[i] = t
	}

	return m
}

func (m Connection) Init() tea.Cmd {
	return textinput.Blink
}

func (m Connection) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit

		case "ctrl+r":
			m.cursorMode++
			if m.cursorMode > cursor.CursorHide {
				m.cursorMode = cursor.CursorBlink
			}
			cmds := make([]tea.Cmd, len(m.inputs))
			for i := range m.inputs {
				cmds[i] = m.inputs[i].Cursor.SetMode(m.cursorMode)
			}
			return m, tea.Batch(cmds...)

		case "tab", "shift+tab", "enter", "up", "down":
			s := msg.String()

			// connection button: connect to DB
			if s == "enter" && m.focusIndex == len(m.inputs) {
				host := m.inputs[0].Value()
				port := m.inputs[1].Value()
				user := m.inputs[2].Value()
				pass := m.inputs[3].Value()
				dbname := m.inputs[4].Value()

				return m, connectDB(host, port, user, pass, dbname, m.db)
			}

			if s == "up" || s == "shift+tab" {
				m.focusIndex--
			} else {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs) {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = noStyle
				m.inputs[i].TextStyle = noStyle
			}

			return m, tea.Batch(cmds...)
		}
	case connectionResultMsg:
		if msg.err != nil {
			m.status = fmt.Sprintf("Connection failed: %v", msg.err)
			return m, nil
		}

		m.db = msg.db
		m.status = "Connected to Postgres successfully!"
		return m, func() tea.Msg { return ConnectionSuccessMsg{} }

	}

	cmd := m.updateInputs(msg)

	return m, cmd
}

func (m *Connection) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.inputs))

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func connectDB(host, port, user, pass, dbname string, db *sql.DB) tea.Cmd {
	return func() tea.Msg {
		connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, pass, dbname)

		postgres, err := sql.Open("postgres", connStr)
		if err != nil {
			return connectionResultMsg{err: err}
		}

		if err := postgres.Ping(); err != nil {
			return connectionResultMsg{err: err}
		}

		db = postgres

		return connectionResultMsg{err: nil, db: db}
	}
}

type connectionResultMsg struct {
	err error
	db  *sql.DB
}

func (m Connection) View() string {
	var b strings.Builder

	fmt.Fprintf(&b, "\n\nConfigure Postgres connection:\n\n")

	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}

	button := &blurredButton
	if m.focusIndex == len(m.inputs) {
		button = &focusedButton
	}
	fmt.Fprintf(&b, "\n\n%s\n\n", *button)
	if m.status != "" {
		b.WriteString(m.status + "\n\n")
	}

	b.WriteString(helpStyle.Render("cursor mode is "))
	b.WriteString(cursorModeHelpStyle.Render(m.cursorMode.String()))
	b.WriteString(helpStyle.Render(" (ctrl+r to change style)"))

	return b.String()
}
