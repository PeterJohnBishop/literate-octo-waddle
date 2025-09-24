package tui

import (
	"database/sql"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func listRecords(db *sql.DB, table string) ([]map[string]interface{}, error) {
	query := fmt.Sprintf(`SELECT * FROM "%s"`, strings.ReplaceAll(table, `"`, `""`))

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := []map[string]interface{}{}

	for rows.Next() {
		// Create a slice of empty interfaces to hold each column value
		values := make([]interface{}, len(columns))
		// Create a slice of pointers to each value in the slice
		valuePtrs := make([]interface{}, len(columns))

		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := values[i]

			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}

		results = append(results, rowMap)
	}

	return results, nil
}

type Records struct {
	status   string
	table    string
	db       *sql.DB
	records  []map[string]any
	cursor   int
	selected map[int]struct{}
	page     int // current page
	pageSize int // how many rows per page
	columns  []string
}

func InitialRecordsModel() Records {
	return Records{
		status:   "",
		selected: make(map[int]struct{}),
		pageSize: 10,
	}
}

func (m Records) Init() tea.Cmd {
	return nil
}

func (m Records) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "ctrl+b":
			return m, func() tea.Msg { return BackToTablesMsg{} }

		case "down", "j":
			if m.cursor < len(m.records)-1 {
				m.cursor++
				// auto advance page if needed
				if m.cursor >= (m.page+1)*m.pageSize {
					m.page++
				}
			}

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
				// auto go back a page
				if m.cursor < m.page*m.pageSize {
					m.page--
				}
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
		}

	case TableSelectedMsg:
		m.db = msg.db
		m.status = "Loading records..."
		records, err := listRecords(m.db, msg.table)
		if err != nil {
			m.status = fmt.Sprintf("Error fetching records: %v", err)
			return m, nil
		}

		m.records = records

		if len(m.records) > 0 {
			// safe to read from the first record
			m.columns = make([]string, 0, len(m.records[0]))
			for col := range m.records[0] {
				m.columns = append(m.columns, col)
			}
			m.status = fmt.Sprintf(
				"Found %d records. Use arrow keys to navigate, space to select.",
				len(m.records),
			)
		} else {
			m.columns = []string{} // reset columns if empty
			m.status = "No records found in the database."
		}

	case BackToConnectionMsg:
		return InitialTablesModel(), nil
	}

	return m, nil
}

func formatRecord(record map[string]any, columns []string) string {
	values := make([]string, len(columns))
	for i, col := range columns {
		if v, ok := record[col]; ok && v != nil {
			values[i] = fmt.Sprintf("%v", v)
		} else {
			values[i] = "NULL"
		}
	}
	return strings.Join(values, " | ")
}

func (m Records) View() string {
	if m.db == nil {
		return ""
	}

	if len(m.records) == 0 {
		return "\n\nNo records found in the database.\n\nDouble check your connection, q to quit.\n"
	}

	start := m.page * m.pageSize
	end := start + m.pageSize
	if end > len(m.records) {
		end = len(m.records)
	}

	s := "\n\nRecords:\n\n"
	// print column headers
	if len(m.columns) > 0 {
		s += strings.Join(m.columns, " | ") + "\n"
		s += strings.Repeat("-", len(s)) + "\n"
	}

	for i := start; i < end; i++ {
		record := m.records[i]

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, formatRecord(record, m.columns))
	}

	s += fmt.Sprintf("\nPage %d of %d | %s\n", m.page+1, (len(m.records)+m.pageSize-1)/m.pageSize, m.status)
	s += "Use ↑/↓ to navigate, space to select, q to quit.\n"

	return s
}
