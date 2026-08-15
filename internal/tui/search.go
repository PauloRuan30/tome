package tui

import tea "github.com/charmbracelet/bubbletea"

type SearchModel struct{ query string }

func NewSearchModel() SearchModel    { return SearchModel{} }
func (m *SearchModel) Query() string { return m.query }
func (m *SearchModel) Reset()        { m.query = "" }

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.String() {
		case "backspace":
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
			}
		case "space":
			m.query += " "
		default:
			if len(key.String()) == 1 {
				m.query += key.String()
			}
		}
	}
	return m, nil
}
