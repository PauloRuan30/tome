package pdf

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// Out-of-band Kitty graphics painter.
//
// Bubble Tea's diffing renderer strips \x1b_G… (APC) sequences, so images
// are written directly to the controlling terminal AFTER each frame lands.
// This is the same technique production TUIs (yazi, ueberzugpp) use.

var (
	ttyOnce sync.Once
	ttyFile *os.File
	ttyErr  error
	wmu     sync.Mutex
)

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

// PaintImage displays a JPEG at 0-based cell (row,col), spanning cols×rows
// cells. Payload is chunked per the Kitty spec (4096-byte chunks, q=2 silent).
func PaintImage(imgData []byte, row, col, cols, rows int) error {
	f, err := tty()
	if err != nil {
		return err
	}
	b64 := base64.StdEncoding.EncodeToString(imgData)

	wmu.Lock()
	defer wmu.Unlock()

	fmt.Fprintf(f, "\x1b[%d;%dH", row+1, col+1) // CUP is 1-based

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
	return nil
}

// PaintVerbose is identical to PaintImage but uses q=1, so Kitty replies
// with "OK" or a concrete error code. Protocol-level debugging only.
func PaintVerbose(w io.Writer, imgData []byte, row, col, cols, rows int) {
	b64 := base64.StdEncoding.EncodeToString(imgData)
	fmt.Fprintf(w, "\x1b[%d;%dH", row+1, col+1)

	const chunk = 4096
	for i := 0; i < len(b64); i += chunk {
		end := min(i+chunk, len(b64))
		m := 0
		if end < len(b64) {
			m = 1
		}
		if i == 0 {
			fmt.Fprintf(w, "\x1b_Ga=T,f=100,c=%d,r=%d,q=1,m=%d;%s\x1b\\", cols, rows, m, b64[i:end])
		} else {
			fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", m, b64[i:end])
		}
	}
}

func Placeholder(cols, rows int) string {
	line := strings.Repeat(" ", cols)
	lines := make([]string, rows)
	for i := range lines {
		lines[i] = line
	}
	return strings.Join(lines, "\n")
}
