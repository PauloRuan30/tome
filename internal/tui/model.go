package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/pdf"
	"github.com/PauloRuan30/tome/internal/progress"
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
		var c1, c2, c3 tea.Cmd
		m.grid, c1 = m.grid.Update(msg)
		m.pane, c2 = m.pane.Update(msg)
		m.reader, c3 = m.reader.Update(msg)
		return m, tea.Batch(c1, c2, c3)

	case ReaderClosedMsg:
		m.mode = modeGrid
		m.grid.InvalidatePaint() // force cover repaint
		m.pane.SetBook(m.currentBooks()[m.selected])
		return m, nil

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
			case "esc": // cancel: restore full library
				m.filter = ""
				m.applyFilter()
				m.mode = modeGrid
				return m, nil
			case "enter": // apply: keep filter, close overlay
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
		m.filter = m.search.Query() // live filtering while typing
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

	var bar string
	barStyle := lipgloss.NewStyle().Background(lipgloss.Color("236")).Foreground(lipgloss.Color("255"))
	if m.mode == modeSearch {
		bar = barStyle.Render(" /search: " + m.search.Query() + "█   (enter=apply  esc=clear)")
	} else {
		bar = barStyle.Render(fmt.Sprintf(
			" term:%dx%d (%s) | sel:%d | scroll:%d | books:%d/%d | filter:%q ",
			m.width, m.height, os.Getenv("TERM"),
			m.selected, m.grid.scrollRow,
			len(m.currentBooks()), len(m.allBooks), m.filter))
	}
	return lipgloss.JoinVertical(lipgloss.Left, body, bar)
}
