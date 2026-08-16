package pdf

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	ttyOnce sync.Once
	ttyFile *os.File
	ttyErr  error
	wmu     sync.Mutex
)

type PageSpec struct {
	Data       []byte
	Row, Col   int
	Cols, Rows int
}

func tty() (*os.File, error) {
	ttyOnce.Do(func() {
		ttyFile, ttyErr = os.OpenFile("/dev/tty", os.O_WRONLY, 0)
	})
	return ttyFile, ttyErr
}

// ClearAllImages deletes every Kitty placement on screen.
func ClearAllImages() {
	f, err := tty()
	if err != nil {
		return
	}
	wmu.Lock()
	defer wmu.Unlock()
	fmt.Fprint(f, "\x1b_Ga=d;\x1b\\")
}

// Repaint atomically clears ALL placements and draws the specs in one
// locked batch — no interleaving races between concurrent paint cmds.
func Repaint(specs []PageSpec) {
	f, err := tty()
	if err != nil {
		return
	}
	wmu.Lock()
	defer wmu.Unlock()

	fmt.Fprint(f, "\x1b_Ga=d;\x1b\\")
	for _, s := range specs {
		paintLocked(f, s.Data, s.Row, s.Col, s.Cols, s.Rows)
	}
}

// PaintImage displays a single image (kept for one-shot tests).
func PaintImage(imgData []byte, row, col, cols, rows int) error {
	f, err := tty()
	if err != nil {
		return err
	}
	wmu.Lock()
	defer wmu.Unlock()
	paintLocked(f, imgData, row, col, cols, rows)
	return nil
}

func paintLocked(f *os.File, imgData []byte, row, col, cols, rows int) {
	b64 := base64.StdEncoding.EncodeToString(imgData)
	fmt.Fprintf(f, "\x1b[%d;%dH", row+1, col+1)

	const chunk = 4096
	for i := 0; i < len(b64); i += chunk {
		end := min(i+chunk, len(b64))
		m := 0
		if end < len(b64) {
			m = 1
		}
		if i == 0 {
			fmt.Fprintf(f, "\x1b_Ga=T,f=100,c=%d,r=%d,q=2,m=%d;%s\x1b\\", cols, rows, m, b64[i:end])
		} else {
			fmt.Fprintf(f, "\x1b_Gm=%d;%s\x1b\\", m, b64[i:end])
		}
	}
}

// Placeholder reserves a cols×rows block of spaces in the TUI layout.
func Placeholder(cols, rows int) string {
	line := strings.Repeat(" ", cols)
	lines := make([]string, rows)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
