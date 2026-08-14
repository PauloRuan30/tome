package tui

import "github.com/charmbracelet/lipgloss"

// Our color palette and base styles
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212"))
	PaneStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(lipgloss.Color("240"))

	CellStyle       = lipgloss.NewStyle().Padding(1, 2)
	ActiveCellStyle = CellStyle.Copy().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("212")) // Highlight selected
)
