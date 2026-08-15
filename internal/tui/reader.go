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

type ReaderClosedMsg struct{}

type readMode int

const (
	readText readMode = iota
	readVisual
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

	lines      []string
	scrollLine int
	textCache  map[int][]string
	pageCache  map[string]string
	paintSig   string
}

func NewReaderModel(t *progress.Tracker) ReaderModel {
	return ReaderModel{tracker: t, pageCache: map[string]string{}, textCache: map[int][]string{}}
}

func (m ReaderModel) Open(book Book) ReaderModel {
	m.book = book
	m.pageCache = map[string]string{}
	m.textCache = map[int][]string{}
	m.jumpMode = false
	m.jumpBuf = ""
	m.paintSig = ""
	// Pixel terminals default to Okular mode; the rest get readable text.
	if pdf.SupportsKitty() {
		m.mode = readVisual
	} else {
		m.mode = readText
	}

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

// visualGeom: deterministic layout shared by View AND the painter.
// row0=header, row1=blank, rows 2..2+rows-1 = page, then footer.
func (m ReaderModel) visualGeom() (rows, cols, startCol, startRow int) {
	rows = m.height - 5
	if rows < 5 {
		rows = 5
	}
	cols = int(float64(rows) / 0.7)
	if m.dual {
		cols = min(cols, (m.width-8)/2)
	} else {
		cols = min(cols, m.width-8)
	}
	if cols < 20 {
		cols = 20
	}
	contentW := cols
	if m.dual {
		contentW = cols*2 + 2
	}
	startCol = (m.width - contentW) / 2
	if startCol < 0 {
		startCol = 0
	}
	return rows, cols, startCol, 2
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

// paintCmd renders the page(s) at 150 DPI and paints them after the frame.
func (m ReaderModel) paintCmd() tea.Cmd {
	file := m.book.Info.FilePath
	page, dual := m.page, m.dual
	rows, cols, startCol, startRow := m.visualGeom()
	return func() tea.Msg {
		time.Sleep(25 * time.Millisecond)
		pdf.ClearAllImages()
		if buf, err := pdf.RenderPage(file, page); err == nil && buf != nil {
			_ = pdf.PaintImage(buf.Bytes(), startRow, startCol, cols, rows)
		}
		if dual {
			if buf, err := pdf.RenderPage(file, page+1); err == nil && buf != nil {
				_ = pdf.PaintImage(buf.Bytes(), startRow, startCol+cols+2, cols, rows)
			}
		}
		return nil
	}
}

func indent(block string, n int) string {
	pad := strings.Repeat(" ", n)
	lines := strings.Split(block, "\n")
	for i, l := range lines {
		lines[i] = pad + l
	}
	return strings.Join(lines, "\n")
}

func joinSideBySide(left, right string, startCol int) string {
	ls, rs := strings.Split(left, "\n"), strings.Split(right, "\n")
	pad := strings.Repeat(" ", startCol)
	var sb strings.Builder
	for i := 0; i < max(len(ls), len(rs)); i++ {
		l, r := "", ""
		if i < len(ls) {
			l = ls[i]
		}
		if i < len(rs) {
			r = rs[i]
		}
		sb.WriteString(pad + l + "  " + r + "\n")
	}
	return strings.TrimSuffix(sb.String(), "\n")
}

func (m ReaderModel) Update(msg tea.Msg) (ReaderModel, tea.Cmd) {
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		if ws.Width != m.width || ws.Height != m.height {
			m.width, m.height = ws.Width, ws.Height
			m.textCache = map[int][]string{}
			m.pageCache = map[string]string{}
			if m.book.Info != nil && m.mode == readText {
				m.loadText()
			}
		}
	} else if ms, ok := msg.(tea.MouseMsg); ok && ms.Action == tea.MouseActionPress {
		switch ms.Button {
		case tea.MouseButtonWheelDown:
			if m.mode == readText {
				m.scrollLine += 3
			} else {
				m.gotoPage(m.page + m.stepDelta())
			}
		case tea.MouseButtonWheelUp:
			if m.mode == readText {
				m.scrollLine -= 3
			} else {
				m.gotoPage(m.page - m.stepDelta())
			}
		}
	} else if key, ok := msg.(tea.KeyMsg); ok {
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
		} else {
			switch key.String() {
			case "q", "esc":
				m.save()
				pdf.ClearAllImages()
				return m, func() tea.Msg { return ReaderClosedMsg{} }
			case "v":
				if m.mode == readText {
					m.mode = readVisual
				} else {
					m.mode = readText
					m.loadText()
				}
			case "d":
				m.dual = !m.dual
				m.mode = readVisual
				m.pageCache = map[string]string{}
			case ":":
				m.jumpMode, m.jumpBuf = true, ""
			case "j", "down":
				if m.mode == readText {
					m.scrollLine++
				} else {
					m.gotoPage(m.page + m.stepDelta())
				}
			case "k", "up":
				if m.mode == readText {
					m.scrollLine--
				} else {
					m.gotoPage(m.page - m.stepDelta())
				}
			case "n", "right", " ", "enter":
				if m.mode == readText {
					m.gotoPage(m.page + 1)
				} else {
					m.gotoPage(m.page + m.stepDelta())
				}
			case "p", "left":
				if m.mode == readText {
					m.gotoPage(m.page - 1)
				} else {
					m.gotoPage(m.page - m.stepDelta())
				}
			}
		}
	}

	// Repaint pixels whenever page/geometry changes (Kitty only)
	var cmd tea.Cmd
	if pdf.SupportsKitty() && m.book.Info != nil && m.mode == readVisual {
		sig := fmt.Sprintf("%d|%t|%d|%d", m.page, m.dual, m.width, m.height)
		if sig != m.paintSig {
			m.paintSig = sig
			cmd = m.paintCmd()
		}
	}
	return m, cmd
}

func (m ReaderModel) stepDelta() int {
	if m.dual {
		return 2
	}
	return 1
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
	header := lipgloss.Place(m.width, 1, lipgloss.Center, lipgloss.Top,
		TitleStyle.Render(fmt.Sprintf("[%s] %s — p.%d/%d (%d%%)",
			tag, m.book.Info.Title, m.page+1, m.book.Info.PageCount, pct)))

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
		m.scrollLine = min(max(m.scrollLine, 0), maxScroll)
		end := min(m.scrollLine+visible, len(m.lines))
		content = strings.Join(m.lines[m.scrollLine:end], "\n")
		footer = fmt.Sprintf("[j/k|wheel] scroll  [n/p|←/→] page  [v] visual  [:] jump  [q] back  · line %d/%d",
			m.scrollLine+1, len(m.lines))
	} else {
		rows, cols, startCol, _ := m.visualGeom()
		hasSecond := m.dual && m.page+1 < m.book.Info.PageCount
		if pdf.SupportsKitty() {
			// Stable placeholders; pixels arrive out-of-band.
			if hasSecond {
				content = joinSideBySide(pdf.Placeholder(cols, rows), pdf.Placeholder(cols, rows), startCol)
			} else {
				content = indent(pdf.Placeholder(cols, rows), startCol)
			}
		} else {
			if hasSecond {
				content = joinSideBySide(m.pageArt(m.page, cols), m.pageArt(m.page+1, cols), startCol)
			} else {
				content = indent(m.pageArt(m.page, cols), startCol)
			}
		}
		footer = "[j/k|←/→|wheel] page  [d] dual  [v] text  [:] jump  [q] back"
	}

	if m.jumpMode {
		footer = "Jump to page: " + m.jumpBuf + "█  (enter=go esc=cancel)"
	}

	return strings.Join([]string{header, "", content, "", footer}, "\n")
}
