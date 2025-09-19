package tui

import (
	"database/sql"

	tea "github.com/charmbracelet/bubbletea"
)

// Which view is active
type State int

const (
	ConnectionView State = iota
	TablesView
)

// Root holds the global db + current sub-model
type Root struct {
	db        *sql.DB
	state     State
	connModel Connection
	tables    Tables
}

// Construct the root with an initial empty connection form
func InitialRootModel(db *sql.DB) Root {
	return Root{
		db:        db,
		state:     ConnectionView,
		connModel: InitialConnectionModel(db), // user fills this in
	}
}

func (r Root) Init() tea.Cmd {
	return nil
}

func (r Root) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch r.state {

	case ConnectionView:
		newModel, cmd := r.connModel.Update(msg)
		if conn, ok := newModel.(Connection); ok {
			r.connModel = conn
		}

		// Look for a successful connection
		switch msg.(type) {
		case ConnectionSuccessMsg:
			r.db = r.connModel.db
			r.tables = InitialTablesModel(r.db)
			r.state = TablesView
			return r, nil
		}

		return r, cmd

	case TablesView:
		newModel, cmd := r.tables.Update(msg)
		if tbls, ok := newModel.(Tables); ok {
			r.tables = tbls
		}

		switch msg.(type) {
		case GoBackMsg:
			r.state = ConnectionView
			return r, nil
		}

		return r, cmd
	}

	return r, nil
}

func (r Root) View() string {
	switch r.state {
	case ConnectionView:
		return r.connModel.View()
	case TablesView:
		return r.tables.View()
	default:
		return "unknown state"
	}
}
