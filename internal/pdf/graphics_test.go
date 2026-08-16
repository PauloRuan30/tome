package pdf

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestANSIBlockArt(t *testing.T) {
	// Create a simple 2x2 image with distinct colors
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255}) // Red
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)

	art := ANSIBlockArt(buf.Bytes(), 5)
	if art == "" || art == "[Image Decode Error]" {
		t.Error("ANSIBlockArt failed to render valid image")
	}
}

func TestSupportsKitty(t *testing.T) {
	_ = SupportsKitty()
}
