package pdf

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"sync"
)

var (
	kittyOnce      sync.Once
	kittySupported bool
)

// SupportsKitty checks if the terminal supports the Kitty Graphics Protocol.
// The result is cached after the first call for performance.
func SupportsKitty() bool {
	kittyOnce.Do(func() {
		term := os.Getenv("TERM")
		termProgram := os.Getenv("TERM_PROGRAM")
		kittySupported = strings.Contains(term, "kitty") ||
			strings.Contains(termProgram, "kitty") ||
			strings.Contains(termProgram, "wezterm")
	})
	return kittySupported
}

// ANSIBlockArt generates high-density Unicode block art for standard terminals.
func ANSIBlockArt(imgData []byte, cols int) string {
	img, _, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return "[Image Decode Error]"
	}

	bounds := img.Bounds()
	aspect := float64(bounds.Dy()) / float64(bounds.Dx())
	rows := int(float64(cols) * aspect * 0.5)
	if rows < 1 {
		rows = 1
	}

	// Pre-allocate builder capacity to reduce GC pressure
	var sb strings.Builder
	sb.Grow(rows * cols * 30)

	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			origX := x * bounds.Dx() / cols
			origY1 := (y * 2) * bounds.Dy() / (rows * 2)
			origY2 := (y*2 + 1) * bounds.Dy() / (rows * 2)

			r1, g1, b1, _ := img.At(origX, origY1).RGBA()
			r2, g2, b2, _ := img.At(origX, origY2).RGBA()

			// Shift 16-bit color to 8-bit
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm▀", r1, g1, b1, r2, g2, b2)
		}
		sb.WriteString("\x1b[0m\n")
	}
	return sb.String()
}
