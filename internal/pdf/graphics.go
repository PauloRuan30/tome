package pdf

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"os"
	"strings"
)

// SupportsKitty detects terminals with the Kitty Graphics Protocol.
// NOTE: Konsole does NOT support it — it correctly uses the ANSI fallback.
func SupportsKitty() bool {
	term := os.Getenv("TERM")
	termProgram := os.Getenv("TERM_PROGRAM")
	return strings.Contains(term, "kitty") ||
		strings.Contains(termProgram, "kitty") ||
		strings.Contains(termProgram, "wezterm")
}

// KittyImage encodes a JPEG per the Kitty Graphics Protocol:
//   - payloads > 4096 bytes MUST be chunked (m=1 ... m=0)
//   - q=2 silences terminal ACK responses (keeps Bubble Tea's input clean)
//   - trailing spaces reserve the layout slot in the TUI
func KittyImage(imgData []byte, cols, rows int) string {
	b64 := base64.StdEncoding.EncodeToString(imgData)

	var sb strings.Builder
	const chunk = 4096
	for i := 0; i < len(b64); i += chunk {
		end := min(i+chunk, len(b64))
		m := 0 // 0 = final chunk
		if end < len(b64) {
			m = 1 // 1 = more chunks follow
		}
		if i == 0 {
			sb.WriteString(fmt.Sprintf(
				"\x1b_Ga=T,f=100,c=%d,r=%d,q=2,m=%d;%s\x1b\\", cols, rows, m, b64[i:end]))
		} else {
			sb.WriteString(fmt.Sprintf("\x1b_Gm=%d;%s\x1b\\", m, b64[i:end]))
		}
	}

	// Layout placeholder so Lipgloss reserves the exact image area
	for y := 0; y < rows; y++ {
		sb.WriteString(strings.Repeat(" ", cols))
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
