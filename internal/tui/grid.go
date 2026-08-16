package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/pdf"
)

type GridModel struct {
	books     []Book
	cols      int
	selected  int
	width     int
	height    int
	scrollRow int

	artCache map[string]string
	paintSig string // last painted viewport signature
}

func NewGridModel(books []Book) GridModel {
	return GridModel{books: books, cols: 3, artCache: make(map[string]string)}
}

func (m GridModel) Selected() int { return m.selected }

// InvalidatePaint forces a repaint on the next update (used after reader exits).
func (m *GridModel) InvalidatePaint() { m.paintSig = "" }

func (m GridModel) coverCols() int {
	c := (m.width / m.cols) - 8
	if c < 20 {
		c = 20
	}
	if c > 40 {
		c = 40
	}
	return c
}
func (m GridModel) coverRows() int { return int(float64(m.coverCols()) * 0.7) }
func (m GridModel) cellH() int     { return m.coverRows() + 5 }
func (m GridModel) totalRows() int { return (len(m.books) + m.cols - 1) / m.cols }
func (m GridModel) visibleRows() int {
	v := m.height / m.cellH()
	if v < 1 {
		v = 1
	}
	return v
}
func (m GridModel) maxScroll() int {
	max := m.totalRows() - m.visibleRows()
	if max < 0 {
		max = 0
	}
	return max
}
func (m GridModel) clampScroll() GridModel {
	if m.scrollRow < 0 {
		m.scrollRow = 0
	}
	if m.scrollRow > m.maxScroll() {
		m.scrollRow = m.maxScroll()
	}
	return m
}
func (m GridModel) keepSelectedVisible() GridModel {
	row := m.selected / m.cols
	if row < m.scrollRow {
		m.scrollRow = row
	}
	if row >= m.scrollRow+m.visibleRows() {
		m.scrollRow = row - m.visibleRows() + 1
	}
	return m.clampScroll()
}

// SetBooks replaces the visible library (search filter).
func (m *GridModel) SetBooks(books []Book) {
	m.books = books
	m.scrollRow = 0
	if m.selected >= len(m.books) {
		m.selected = len(m.books) - 1
	}
	if m.selected < 0 {
		m.selected = 0
	}
}

// coverArt: ANSI art for classic terminals; stable space placeholder for
// Kitty (the real image is painted out-of-band by paintCmd).
func (m GridModel) coverArt(book Book, cols int) string {
	if pdf.SupportsKitty() {
		return pdf.Placeholder(cols, m.coverRows())
	}
	key := fmt.Sprintf("%s-%d", book.Info.Hash, cols)
	if art, ok := m.artCache[key]; ok {
		return art
	}
	var art string
	if book.Cover == nil {
		art = lipgloss.Place(cols, m.coverRows(), lipgloss.Center, lipgloss.Center, "[No Cover]")
	} else {
		art = pdf.ANSIBlockArt(book.Cover, cols)
	}
	if len(m.artCache) > 64 {
		m.artCache = make(map[string]string)
	}
	m.artCache[key] = art
	return art
}

// paintCmd snapshots the visible viewport and paints covers after the frame.
func (m GridModel) paintCmd() tea.Cmd {
	type slot struct {
		cover    []byte
		row, col int
		c, r     int
	}
	var slots []slot
	cellW := m.width / m.cols
	first := m.scrollRow * m.cols
	last := min((m.scrollRow+m.visibleRows())*m.cols, len(m.books))
	for i := first; i < last; i++ {
		if m.books[i].Cover == nil {
			continue
		}
		slots = append(slots, slot{
			cover: m.books[i].Cover,
			row:   (i/m.cols-m.scrollRow)*m.cellH() + 2,
			col:   (i%m.cols)*cellW + 3,
			c:     m.coverCols(),
			r:     m.coverRows(),
		})
	}
	return func() tea.Msg {
		time.Sleep(25 * time.Millisecond)
		specs := make([]pdf.PageSpec, 0, len(slots))
		for _, s := range slots {
			specs = append(specs, pdf.PageSpec{
				Data: s.cover,
				Row:  s.row,
				Col:  s.col,
				Cols: s.c,
				Rows: s.r,
			})
		}
		pdf.Repaint(specs)
		return nil
	}
}

func (m GridModel) Update(msg tea.Msg) (GridModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
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
		m = m.clampScroll()

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
		m = m.keepSelectedVisible()

	case tea.MouseMsg:
		if msg.Action != tea.MouseActionPress {
			return m, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelDown:
			m.scrollRow++
		case tea.MouseButtonWheelUp:
			m.scrollRow--
		case tea.MouseButtonLeft:
			if msg.X >= m.width {
				return m, nil
			}
			col := msg.X / (m.width / m.cols)
			if col >= m.cols {
				return m, nil
			}
			row := m.scrollRow + msg.Y/m.cellH()
			idx := row*m.cols + col
			if idx >= 0 && idx < len(m.books) {
				m.selected = idx
			}
		}
		m = m.clampScroll()
	}

	// Repaint covers whenever the visible viewport changes (Kitty only)
	if pdf.SupportsKitty() {
		sig := fmt.Sprintf("%d|%d|%d|%d|%d", m.scrollRow, m.cols, m.width, m.height, len(m.books))
		if sig != m.paintSig {
			m.paintSig = sig
			cmd = m.paintCmd()
		}
	}
	return m, cmd
}

func (m GridModel) View() string {
	if len(m.books) == 0 {
		return "No PDFs found in this directory."
	}
	m = m.clampScroll()
	coverCols := m.coverCols()

	first := m.scrollRow * m.cols
	last := min((m.scrollRow+m.visibleRows())*m.cols, len(m.books))

	var cells []string
	for i := first; i < last; i++ {
		book := m.books[i]
		style := CellStyle
		if i == m.selected {
			style = ActiveCellStyle
		}

		coverView := m.coverArt(book, coverCols)

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

	if m.maxScroll() > 0 {
		grid.WriteString(fmt.Sprintf(" ⇕ row %d/%d", m.scrollRow+1, m.maxScroll()+1))
	}
	return grid.String()
}
