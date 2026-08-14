package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/pdf"
)

type GridModel struct {
	books    []Book
	cols     int
	selected int
	width    int
	height   int
}

func NewGridModel(books []Book) GridModel {
	return GridModel{books: books, cols: 3}
}

func (m GridModel) Selected() int { return m.selected }

// coverCols: how many terminal columns wide each cover art should be.
// Scales with the window => sharper art on bigger terminals.
func (m GridModel) coverCols() int {
	c := (m.width / m.cols) - 8 // reserve space for padding + border
	if c < 20 {
		c = 20
	}
	if c > 60 {
		c = 60
	}
	return c
}

// coverRows: A4 pages have a ~1.414 h/w ratio; a terminal char is ~2x tall
// as wide, so rows ≈ cols * 0.7.
func (m GridModel) coverRows() int {
	return int(float64(m.coverCols()) * 0.7)
}

// cellH: total terminal rows one grid slot occupies (cover + title + chrome).
func (m GridModel) cellH() int {
	return m.coverRows() + 5
}

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Responsive reflow: more width => more columns
		switch {
		case m.width > 140:
			m.cols = 4
		case m.width > 100:
			m.cols = 3
		case m.width > 60:
			m.cols = 2
		default:
			m.cols = 1
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "h", "left":
			if m.selected > 0 {
				m.selected--
			}
		case "l", "right":
			if m.selected < len(m.books)-1 {
				m.selected++
			}
		case "k", "up":
			if m.selected >= m.cols {
				m.selected -= m.cols
			}
		case "j", "down":
			if m.selected+m.cols < len(m.books) {
				m.selected += m.cols
			}
		}

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			if m.selected+m.cols < len(m.books) {
				m.selected += m.cols
			}
		case tea.MouseButtonWheelUp:
			if m.selected >= m.cols {
				m.selected -= m.cols
			}
		case tea.MouseButtonLeft:
			// 1. Ignore clicks that land in the sidebar (right 30%)
			if msg.X >= m.width {
				return m, nil
			}
			// 2. Map X/Y to a grid slot
			cellW := m.width / m.cols
			col := msg.X / cellW
			row := msg.Y / m.cellH()

			// 3. Ignore clicks in the empty gutter at the end of a row
			if col >= m.cols {
				return m, nil
			}

			idx := row*m.cols + col
			// 4. Ignore clicks on empty space below the last book
			if idx >= 0 && idx < len(m.books) {
				m.selected = idx
			}
		}
	}
	return m, nil
}

func (m GridModel) View() string {
	if len(m.books) == 0 {
		return "No PDFs found in this directory."
	}

	coverCols := m.coverCols()
	coverRows := m.coverRows()

	var cells []string
	for i, book := range m.books {
		style := CellStyle
		if i == m.selected {
			style = ActiveCellStyle
		}

		var coverView string
		if book.Cover != nil {
			if pdf.SupportsKitty() {
				coverView = pdf.KittyImage(book.Cover, coverCols, coverRows)
			} else {
				// Higher resolution now: scales with terminal width
				coverView = pdf.ANSIBlockArt(book.Cover, coverCols)
			}
		} else {
			coverView = lipgloss.Place(coverCols, coverRows, lipgloss.Center, lipgloss.Center, "[No Cover]")
		}

		title := book.Info.Title
		if len(title) > coverCols-3 {
			title = title[:coverCols-6] + "..."
		}

		cells = append(cells, style.Render(
			lipgloss.JoinVertical(lipgloss.Center, coverView, title),
		))
	}

	var grid strings.Builder
	for i := 0; i < len(cells); i += m.cols {
		end := i + m.cols
		if end > len(cells) {
			end = len(cells)
		}
		grid.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, cells[i:end]...) + "\n")
	}

	return grid.String()
}
