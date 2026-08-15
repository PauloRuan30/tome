package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PaneModel struct {
	book   Book
	width  int
	height int
}

func NewPaneModel() PaneModel { return PaneModel{} }

func (m *PaneModel) SetBook(book Book) { m.book = book }

func (m PaneModel) Update(msg tea.Msg) (PaneModel, tea.Cmd) { return m, nil }

func (m PaneModel) View() string {
	if m.book.Info == nil {
		return PaneStyle.Width(m.width - 4).Render("No book selected")
	}

	style := PaneStyle.Width(m.width - 4)

	title := TitleStyle.Render(m.book.Info.Title)
	author := fmt.Sprintf("Author: %s", m.book.Info.Author)
	pages := fmt.Sprintf("Pages: %d", m.book.Info.PageCount)
	path := fmt.Sprintf("Path: %s", m.book.Info.FilePath)
	hash := fmt.Sprintf("Hash: %s...", m.book.Info.Hash[:8])

	content := lipgloss.JoinVertical(lipgloss.Left,
		title, "", author, pages, "", path, hash, "",
		"--- Controls ---",
		"[q] Quit",
		"[Space/Enter] Read (Step 5)",
	)
	return style.Render(content)
}
