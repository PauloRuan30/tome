package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/progress"
)

type PaneModel struct {
	book    Book
	tracker *progress.Tracker
	width   int
	height  int
}

func NewPaneModel(t *progress.Tracker) PaneModel { return PaneModel{tracker: t} }
func (m *PaneModel) SetBook(book Book)           { m.book = book }
func (m PaneModel) Update(msg tea.Msg) (PaneModel, tea.Cmd) {
	return m, nil
}

func (m PaneModel) View() string {
	style := PaneStyle.Width(m.width - 4)
	if m.book.Info == nil {
		return style.Render("No book selected")
	}

	progLine := "Progress: never opened"
	lastOpen := "Last opened: -"
	if p := m.tracker.Get(m.book.Info.Hash); p.LastPage > 0 {
		pct := p.LastPage * 100 / max(p.Total, 1)
		progLine = fmt.Sprintf("Progress: page %d/%d (%d%%)", p.LastPage, p.Total, pct)
		lastOpen = "Last opened: " + p.LastOpen.Format("2006-01-02 15:04")
	}

	content := lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render(m.book.Info.Title), "",
		fmt.Sprintf("Author: %s", m.book.Info.Author),
		fmt.Sprintf("Pages: %d", m.book.Info.PageCount), "",
		fmt.Sprintf("Path: %s", m.book.Info.FilePath),
		fmt.Sprintf("Hash: %s...", m.book.Info.Hash[:8]), "",
		progLine,
		lastOpen, "",
		"--- Controls ---",
		"[Enter/Space] Read   [o] External",
		"[/] Search           [q] Quit",
	)
	return style.Render(content)
}
