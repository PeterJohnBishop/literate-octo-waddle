package tui

// this model will act as a global container for the other models
import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Parent struct {
	connection Connection
	tables     Tables
	records    Records
}

type ConnectionSuccessMsg struct {
	db *sql.DB
}

type BackToConnectionMsg struct{}
type TablesLoadedMsg struct{}
type TableSelectedMsg struct {
	table string
	db    *sql.DB
}
type BackToTablesMsg struct{}

func InitialParentModel() Parent {
	return Parent{
		connection: InitialConnectionModel(),
		tables:     InitialTablesModel(),
		records:    InitialRecordsModel(),
	}
}

func (m Parent) Init() tea.Cmd {
	return nil
}

func (m Parent) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "q":
			return m, tea.Quit

		}
	}

	var updated tea.Model
	var cmd tea.Cmd
	updated, cmd = m.connection.Update(msg)
	m.connection = updated.(Connection)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	updated, cmd = m.tables.Update(msg)
	m.tables = updated.(Tables)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	updated, cmd = m.records.Update(msg)
	m.records = updated.(Records)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Parent) View() string {

	row := "\nWelcome to the Postgres TUI Client!\n\n"
	// Row of children
	row += lipgloss.JoinHorizontal(
		lipgloss.Top,
		m.connection.View(),
		m.tables.View(),
	)

	// Bottom child below
	col := lipgloss.JoinVertical(lipgloss.Left,
		row,
		m.records.View(),
	)

	return col
}
