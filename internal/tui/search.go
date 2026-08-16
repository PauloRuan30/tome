package tui

import tea "github.com/charmbracelet/bubbletea"

type SearchModel struct{ query string }

func NewSearchModel() SearchModel    { return SearchModel{} }
func (m *SearchModel) Query() string { return m.query }
func (m *SearchModel) Reset()        { m.query = "" }

func (m SearchModel) Update(msg tea.Msg) (SearchModel, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Type {
		case tea.KeyBackspace:

			runes := []rune(m.query)
			if len(runes) > 0 {
				m.query = string(runes[:len(runes)-1])
			}
		case tea.KeyRunes:
			// Appends all typed or pasted runes
			m.query += string(key.Runes)
		case tea.KeySpace:
			m.query += " "
		}
	}
	return m, nil
}
