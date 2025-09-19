package main

import (
	"database/sql"
	"fmt"
	"os"
	"tui/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var db *sql.DB

func main() {
	if _, err := tea.NewProgram(tui.InitialRootModel(db)).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
