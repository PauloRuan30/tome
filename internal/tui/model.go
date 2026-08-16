package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/PauloRuan30/tome/internal/pdf"

	"github.com/PauloRuan30/tome/internal/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Book struct {
	Info  *pdf.BookInfo
	Cover []byte
}

type viewMode int

const (
	modeGrid viewMode = iota
	modeSearch
	modeReader
)

type Model struct {
	allBooks []Book
	grid     GridModel
	pane     PaneModel
	reader   ReaderModel
	search   SearchModel
	tracker  *progress.Tracker
	mode     viewMode
	width    int
	height   int
	selected int
	filter   string
}

func InitialModel(books []Book, tracker *progress.Tracker) Model {
	m := Model{
		allBooks: books,
		grid:     NewGridModel(books),
		pane:     NewPaneModel(tracker),
		reader:   NewReaderModel(tracker),
		search:   NewSearchModel(),
		tracker:  tracker,
	}
	if len(books) > 0 {
		m.pane.SetBook(books[0])
	}
	return m
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		gridWidth := int(float64(msg.Width) * 0.70)
		m.grid.width, m.grid.height = gridWidth, msg.Height-1
		m.pane.width, m.pane.height = msg.Width-gridWidth, msg.Height-1
		m.reader.width, m.reader.height = msg.Width, msg.Height

		var cmds []tea.Cmd
		if m.mode == modeReader {
			var c tea.Cmd
			m.reader, c = m.reader.Update(msg)
			cmds = append(cmds, c)
		} else {
			var c1, c2 tea.Cmd
			m.grid, c1 = m.grid.Update(msg)
			m.pane, c2 = m.pane.Update(msg)
			cmds = append(cmds, c1, c2)
		}
		return m, tea.Batch(cmds...)

	case ReaderClosedMsg:
		m.mode = modeGrid
		m.grid.InvalidatePaint()
		var cmd tea.Cmd
		m.grid, cmd = m.grid.Update(nil)
		if books := m.currentBooks(); len(books) > 0 && m.selected < len(books) {
			m.pane.SetBook(books[m.selected])
		}
		return m, cmd

	case tea.KeyMsg:
		if m.mode == modeGrid {
			switch msg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "/":
				m.mode = modeSearch
				m.search.Reset()
				return m, nil
			case "enter", " ":
				if books := m.currentBooks(); len(books) > 0 {
					m.reader = m.reader.Open(books[m.selected])
					m.mode = modeReader
					return m, m.reader.PaintIfNeeded()
				}
				return m, nil
			case "o":
				if books := m.currentBooks(); len(books) > 0 {
					return m, openExternal(books[m.selected])
				}
				return m, nil
			}
		}
		if m.mode == modeSearch {
			switch msg.String() {
			case "esc":
				m.filter = ""
				m.applyFilter()
				m.mode = modeGrid
				return m, nil
			case "enter":
				m.filter = m.search.Query()
				m.applyFilter()
				m.mode = modeGrid
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	switch m.mode {
	case modeGrid:
		m.grid, cmd = m.grid.Update(msg)
		if m.grid.Selected() != m.selected {
			m.selected = m.grid.Selected()
			if books := m.currentBooks(); m.selected < len(books) {
				m.pane.SetBook(books[m.selected])
			}
		}
	case modeSearch:
		m.search, cmd = m.search.Update(msg)
		m.filter = m.search.Query()
		m.applyFilter()
	case modeReader:
		m.reader, cmd = m.reader.Update(msg)
	}
	return m, cmd
}

func (m Model) currentBooks() []Book { return m.grid.books }

func (m *Model) applyFilter() {
	q := strings.ToLower(strings.TrimSpace(m.filter))
	if q == "" {
		m.grid.SetBooks(m.allBooks)
	} else {
		var fb []Book
		for _, b := range m.allBooks {
			if strings.Contains(strings.ToLower(b.Info.Title), q) ||
				strings.Contains(strings.ToLower(b.Info.Author), q) ||
				strings.Contains(strings.ToLower(b.Info.FilePath), q) {
				fb = append(fb, b)
			}
		}
		m.grid.SetBooks(fb)
	}
	m.selected = m.grid.Selected()
	if books := m.currentBooks(); len(books) > 0 && m.selected < len(books) {
		m.pane.SetBook(books[m.selected])
	}
}

func openExternal(b Book) tea.Cmd {
	return func() tea.Msg {
		_ = exec.Command("xdg-open", b.Info.FilePath).Start()
		return nil
	}
}

func (m Model) View() string {
	if m.mode == modeReader {
		return m.reader.View()
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, m.grid.View(), m.pane.View())

	barStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("255")).
		Width(m.width)

	var bar string
	if m.mode == modeSearch {
		bar = barStyle.Render(fmt.Sprintf(" 🔍 Search: %s█   (enter=apply  esc=clear)", m.search.Query()))
	} else {
		status := fmt.Sprintf(" 📚 %d books", len(m.currentBooks()))
		if m.filter != "" {
			status += fmt.Sprintf(" (filtered by: %q)", m.filter)
		}
		status += "  |  Press / to search, q to quit"
		bar = barStyle.Render(status)
	}

	return lipgloss.JoinVertical(lipgloss.Left, body, bar)
}
