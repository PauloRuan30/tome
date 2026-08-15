package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/PauloRuan30/tome/internal/pdf"
	"github.com/PauloRuan30/tome/internal/progress"
)

// ReaderClosedMsg tells the parent Model to return to the grid.
type ReaderClosedMsg struct{}

type readMode int

const (
	readText   readMode = iota // extracted text — actually readable (default)
	readVisual                 // block-art page preview
)

type ReaderModel struct {
	tracker *progress.Tracker
	book    Book
	page    int
	dual    bool
	mode    readMode
	width   int
	height  int

	jumpMode bool
	jumpBuf  string

	// text mode state
	lines      []string
	scrollLine int
	textCache  map[int][]string

	// visual mode state
	pageCache map[string]string
}

func NewReaderModel(t *progress.Tracker) ReaderModel {
	return ReaderModel{tracker: t, pageCache: map[string]string{}, textCache: map[int][]string{}}
}

// Open loads a book and resumes from the last saved page.
func (m ReaderModel) Open(book Book) ReaderModel {
	m.book = book
	m.pageCache = map[string]string{}
	m.textCache = map[int][]string{}
	m.jumpMode = false
	m.jumpBuf = ""
	m.mode = readText
	m.dual = false

	start := 0
	if p := m.tracker.Get(book.Info.Hash); p.Total == book.Info.PageCount && p.LastPage > 0 {
		start = p.LastPage - 1
	}
	m.page = start
	m.loadText()
	m.save()
	return m
}

func (m *ReaderModel) save() {
	if m.book.Info == nil {
		return
	}
	m.tracker.Update(m.book.Info.Hash, progress.Progress{
		LastPage: m.page + 1,
		Total:    m.book.Info.PageCount,
		LastOpen: time.Now(),
	})
}

func (m *ReaderModel) gotoPage(n int) {
	if m.book.Info == nil || n < 0 || n > m.book.Info.PageCount-1 {
		return
	}
	m.page = n
	m.scrollLine = 0
	if m.mode == readText {
		m.loadText()
	}
	m.save()
}

func (m *ReaderModel) loadText() {
	if lines, ok := m.textCache[m.page]; ok {
		m.lines = lines
		return
	}
	raw, err := pdf.PageText(m.book.Info.FilePath, m.page)
	if err != nil || strings.TrimSpace(raw) == "" {
		m.lines = []string{"[no text layer on this page — press v for visual mode]"}
	} else {
		m.lines = wrapText(raw, m.width-8)
	}
	m.textCache[m.page] = m.lines
}

// wrapText reflows raw page text to the terminal width.
func wrapText(s string, width int) []string {
	if width < 10 {
		width = 80
	}
	var out []string
	for _, raw := range strings.Split(s, "\n") {
		raw = strings.TrimSpace(strings.ReplaceAll(raw, "\x00", ""))
		if raw == "" {
			out = append(out, "")
			continue
		}
		line := ""
		for _, w := range strings.Fields(raw) {
			switch {
			case line == "":
				line = w
			case len(line)+1+len(w) <= width:
				line += " " + w
			default:
				out = append(out, line)
				line = w
			}
		}
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

func (m ReaderModel) pageArt(page, cols int) string {
	key := fmt.Sprintf("%d-%d", page, cols)
	if art, ok := m.pageCache[key]; ok {
		return art
	}
	buf, err := pdf.RenderPage(m.book.Info.FilePath, page)
	if err != nil || buf == nil {
		return "[render error]"
	}
	art := pdf.ANSIBlockArt(buf.Bytes(), cols)
	m.pageCache[key] = art
	return art
}

func (m *ReaderModel) step(n int) {
	delta := n
	if m.dual {
		delta = n * 2
	}
	m.gotoPage(m.page + delta)
}

func (m ReaderModel) Update(msg tea.Msg) (ReaderModel, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		if ws.Width != m.width || ws.Height != m.height {
			m.width, m.height = ws.Width, ws.Height
			m.textCache = map[int][]string{} // rewrap for new width
			m.pageCache = map[string]string{}
			if m.book.Info != nil && m.mode == readText {
				m.loadText()
			}
		}
		return m, nil
	}

	if ms, ok := msg.(tea.MouseMsg); ok && ms.Action == tea.MouseActionPress {
		switch ms.Button {
		case tea.MouseButtonWheelDown:
			if m.mode == readText {
				m.scrollLine += 3
			} else {
				m.step(1)
			}
		case tea.MouseButtonWheelUp:
			if m.mode == readText {
				m.scrollLine -= 3
			} else {
				m.step(-1)
			}
		}
		return m, nil
	}

	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	// page-jump prompt
	if m.jumpMode {
		switch key.String() {
		case "esc":
			m.jumpMode, m.jumpBuf = false, ""
		case "enter":
			if n, err := strconv.Atoi(m.jumpBuf); err == nil && n >= 1 && n <= m.book.Info.PageCount {
				m.gotoPage(n - 1)
			}
			m.jumpMode, m.jumpBuf = false, ""
		case "backspace":
			if len(m.jumpBuf) > 0 {
				m.jumpBuf = m.jumpBuf[:len(m.jumpBuf)-1]
			}
		default:
			if len(key.String()) == 1 && key.String() >= "0" && key.String() <= "9" {
				m.jumpBuf += key.String()
			}
		}
		return m, nil
	}

	switch key.String() {
	case "q", "esc":
		m.save()
		return m, func() tea.Msg { return ReaderClosedMsg{} }

	case "v": // toggle TEXT <-> VISUAL
		if m.mode == readText {
			m.mode = readVisual
		} else {
			m.mode = readText
			m.loadText()
		}

	case "d": // dual-page (visual mode)
		m.dual = !m.dual
		m.mode = readVisual
		m.pageCache = map[string]string{}

	case ":":
		m.jumpMode, m.jumpBuf = true, ""

	// scrolling (text) / page flip (visual)
	case "j", "down":
		if m.mode == readText {
			m.scrollLine++
		} else {
			m.step(1)
		}
	case "k", "up":
		if m.mode == readText {
			m.scrollLine--
		} else {
			m.step(-1)
		}

	// page turning
	case "n", "right", " ", "enter":
		if m.mode == readText {
			m.gotoPage(m.page + 1)
		} else {
			m.step(1)
		}
	case "p", "left":
		if m.mode == readText {
			m.gotoPage(m.page - 1)
		} else {
			m.step(-1)
		}
	}
	return m, nil
}

func (m ReaderModel) View() string {
	if m.book.Info == nil {
		return "nothing to read"
	}

	pct := (m.page + 1) * 100 / max(m.book.Info.PageCount, 1)
	tag := "TEXT"
	if m.mode == readVisual {
		tag = "VISUAL"
	}
	header := TitleStyle.Render(fmt.Sprintf("[%s] %s — p.%d/%d (%d%%)",
		tag, m.book.Info.Title, m.page+1, m.book.Info.PageCount, pct))

	var content, footer string

	if m.mode == readText {
		visible := m.height - 4
		if visible < 1 {
			visible = 1
		}
		maxScroll := len(m.lines) - visible
		if maxScroll < 0 {
			maxScroll = 0
		}
		if m.scrollLine < 0 {
			m.scrollLine = 0
		}
		if m.scrollLine > maxScroll {
			m.scrollLine = maxScroll
		}
		end := min(m.scrollLine+visible, len(m.lines))
		content = strings.Join(m.lines[m.scrollLine:end], "\n")
		footer = fmt.Sprintf("[j/k|wheel] scroll  [n/p|←/→] page  [v] visual  [:] jump  [q] back  · line %d/%d",
			m.scrollLine+1, len(m.lines))
	} else {
		rows := m.height - 4
		cols := int(float64(rows) / 0.7)
		if m.dual {
			cols = min(cols, (m.width/2)-6)
		} else {
			cols = min(cols, m.width-4)
		}
		if cols < 20 {
			cols = 20
		}
		if m.dual && m.page+1 < m.book.Info.PageCount {
			content = lipgloss.JoinHorizontal(lipgloss.Top,
				m.pageArt(m.page, cols), "  ", m.pageArt(m.page+1, cols))
		} else {
			content = m.pageArt(m.page, cols)
		}
		footer = "[j/k|←/→|wheel] page  [d] dual  [v] text  [:] jump  [q] back"
	}

	if m.jumpMode {
		footer = "Jump to page: " + m.jumpBuf + "█  (enter=go esc=cancel)"
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center, header, content, footer))
}
