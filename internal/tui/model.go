package tui

import (
	"github.com/PauloRuan30/tome/internal/pdf"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Book represents a fully parsed book ready for the UI
type Book struct {
	Info  *pdf.BookInfo
	Cover []byte
}

type Model struct {
	books    []Book
	grid     GridModel
	width    int
	height   int
	selected int
}

func InitialModel(books []Book) Model {
	return Model{
		books: books,
		grid:  NewGridModel(books),
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// 70% width for grid, 30% for sidebar
		m.grid.width = int(float64(msg.Width) * 0.7)
		m.grid.height = msg.Height
		m.grid, _ = m.grid.Update(msg)
		return m, nil

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.grid, cmd = m.grid.Update(msg)

	// Sync selected index from grid to main model
	m.selected = m.grid.Selected()

	return m, cmd
}

func (m Model) View() string {
	// Left Pane: Grid
	gridView := m.grid.View()

	// Right Pane: Placeholder Sidebar (We will build this in Step 4)
	sidebarWidth := m.width - m.grid.width
	sidebar := PaneStyle.Width(sidebarWidth - 4).Render(
		TitleStyle.Render("Metadata Sidebar") + "\n\n" +
			"Select a book to see details.\n\n" +
			"[q] Quit",
	)

	// Join them horizontally (Split Screen)
	return lipgloss.JoinHorizontal(lipgloss.Top, gridView, sidebar)
}
