package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"os"
	"strings"
)

// checks if the terminal emulator supports the Kitty Graphics Protocol.
func SupportsKitty() bool {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	return strings.Contains(term, "kitty") ||
		strings.Contains(termProgram, "kitty") ||
		strings.Contains(termProgram, "wezterm") ||
		strings.Contains(term, "konsole")
}

// KittyImage formats a JPEG into the Kitty Graphics Protocol ANSI escape sequence.
func KittyImage(imgData []byte, cols, rows int) string {
	b64 := base64.StdEncoding.EncodeToString(imgData)

	// a=T (display and store), f=100 (JPEG), c=cols, r=rows
	seq := fmt.Sprintf("\x1b_Ga=T,f=100,c=%d,r=%d;%s\x1b\\", cols, rows, b64)

	var sb strings.Builder
	sb.WriteString(seq)

	// We must print placeholder spaces so the TUI layout engine knows how much space the image takes
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			sb.WriteString(" ")
		}
		if y < rows-1 {
			sb.WriteString("\n")
		}
	}
	return sb.String()
}

// ANSIBlockArt generates high-density Unicode block art for standard terminals.
func ANSIBlockArt(imgData []byte, cols int) string {
	img, err := jpeg.Decode(bytes.NewReader(imgData))
	if err != nil {
		return "[Image Decode Error]"
	}

	bounds := img.Bounds()
	aspect := float64(bounds.Dy()) / float64(bounds.Dx())

	rows := int(float64(cols) * aspect * 0.5)
	if rows < 1 {
		rows = 1
	}

	var sb strings.Builder
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			// Map terminal grid (x,y) to image pixel coordinates
			origX := x * bounds.Dx() / cols

			origY1 := (y * 2) * bounds.Dy() / (rows * 2)
			origY2 := (y*2 + 1) * bounds.Dy() / (rows * 2)

			c1 := img.At(origX, origY1)
			c2 := img.At(origX, origY2)

			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()

			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1, r2, g2, b2))
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
