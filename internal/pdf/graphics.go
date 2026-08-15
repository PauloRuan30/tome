package pdf

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg" // Keep for potential external images
	_ "image/png"  // Register PNG decoder
	"os"
	"strings"
)

func SupportsKitty() bool {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	return strings.Contains(term, "kitty") ||
		strings.Contains(termProgram, "kitty") ||
		strings.Contains(termProgram, "wezterm")
}

// ANSIBlockArt uses image.Decode which auto-detects the format (PNG/JPEG).
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

	var sb strings.Builder
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			origX := x * bounds.Dx() / cols
			origY1 := (y * 2) * bounds.Dy() / (rows * 2)
			origY2 := (y*2 + 1) * bounds.Dy() / (rows * 2)

			c1 := img.At(origX, origY1)
			c2 := img.At(origX, origY2)

			r1, g1, b1, _ := c1.RGBA()
			r2, g2, b2, _ := c2.RGBA()
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%d;48;2;%d;%d;%dm▀", r1, g1, b1, r2, g2, b2))
		}
		sb.WriteString("\x1b[0m\n")
	}
	return sb.String()
}
