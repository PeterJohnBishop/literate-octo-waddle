package tui

import (
	"database/sql"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func listTables(db *sql.DB) ([]string, error) {

	rows, err := db.Query(`
		SELECT tablename
		FROM pg_catalog.pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema');
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}

	return tables, nil
}

type Tables struct {
	status   string
	db       *sql.DB
	tables   []string
	cursor   int
	selected map[int]struct{}
}

func InitialTablesModel() Tables {

	return Tables{
		status:   "Waiting for database connection...",
		selected: make(map[int]struct{}),
	}
}

func (m Tables) Init() tea.Cmd {
	return nil
}

func (m Tables) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+b":
			return m, func() tea.Msg { return BackToConnectionMsg{} }

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.tables)-1 {
				m.cursor++
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

	case ConnectionSuccessMsg:
		m.db = msg.db
		m.status = "Connected to database. Loading tables..."
		tables, err := listTables(m.db)
		if err != nil {
			m.status = fmt.Sprintf("Error fetching tables: %v", err)
			return m, nil
		}
		m.tables = tables
		if len(m.tables) == 0 {
			m.status = "No tables found in the database."
		} else {
			m.status = fmt.Sprintf("Found %d tables. Use arrow keys to navigate, space to select.", len(m.tables))
		}

	case BackToConnectionMsg:
		return InitialConnectionModel(), nil
	}

	return m, nil
}

func (m Tables) View() string {

	if m.db == nil {
		return ""
	}

	if len(m.tables) == 0 {
		return "\n\nNo tables found in the database.\n\nDouble check your connection, q to quit.\n"
	}

	s := "\n\nTables: \n\n"

	for i, choice := range m.tables {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
