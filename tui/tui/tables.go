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

type GoBackMsg struct{}

func InitialTablesModel(db *sql.DB) Tables {

	tables, err := listTables(db)
	if err != nil {
		return Tables{
			status: fmt.Sprintf("Error fetching tables: %v", err),
			db:     db,
		}
	}

	return Tables{
		db:       db,
		tables:   tables,
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
			return m, func() tea.Msg { return GoBackMsg{} }

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
	}

	return m, nil
}

func (m Tables) View() string {
	s := "Tables: \n\n"

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

	s += "\nPress ctrl+b for back, q to quit.\n"

	return s
}
