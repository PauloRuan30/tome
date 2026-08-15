package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/pdf"
)

type Book struct {
	Info  *pdf.BookInfo
	Cover []byte
}

type Model struct {
	books    []Book
	grid     GridModel
	pane     PaneModel
	width    int
	height   int
	selected int
}

func InitialModel(books []Book) Model {
	m := Model{books: books, grid: NewGridModel(books), pane: NewPaneModel()}
	if len(books) > 0 {
		m.pane.SetBook(books[0])
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		gridWidth := int(float64(msg.Width) * 0.70)
		m.grid.width = gridWidth
		m.grid.height = msg.Height - 1 // reserve 1 line for debug bar
		m.pane.width = msg.Width - gridWidth
		m.pane.height = msg.Height - 1

		var c1, c2 tea.Cmd
		m.grid, c1 = m.grid.Update(msg)
		m.pane, c2 = m.pane.Update(msg)
		return m, tea.Batch(c1, c2)

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.grid, cmd = m.grid.Update(msg)

	if m.grid.Selected() != m.selected {
		m.selected = m.grid.Selected()
		if m.selected >= 0 && m.selected < len(m.books) {
			m.pane.SetBook(m.books[m.selected])
		}
	}
	return m, cmd
}

func (m Model) View() string {
	body := lipgloss.JoinHorizontal(lipgloss.Top, m.grid.View(), m.pane.View())

	// 🔍 DEBUG BAR — tells us the internal state in every screenshot
	status := fmt.Sprintf(
		" term:%dx%d | gridW:%d paneW:%d | cols:%d | sel:%d | scroll:%d | books:%d ",
		m.width, m.height, m.grid.width, m.pane.width,
		m.grid.cols, m.selected, m.grid.scrollRow, len(m.books),
	)
	bar := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("255")).
		Render(status)

	return lipgloss.JoinVertical(lipgloss.Left, body, bar)
}
