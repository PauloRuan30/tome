package pdf

import (
	"bytes"
	"image/jpeg"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

// BookInfo holds the extracted metadata of a PDF.
type BookInfo struct {
	Title     string
	Author    string
	PageCount int
	FilePath  string
	Hash      string
}

// Parse opens a PDF, extracts metadata and the first page as a JPEG.
func Parse(filePath string, hash string) (*BookInfo, []byte, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer doc.Close()

	cleanStrings := func(s string) string {
		return strings.TrimSpace(strings.ReplaceAll(s, "\x00", ""))
	}

	info := &BookInfo{
		Title:     cleanStrings(doc.Metadata()["title"]),
		Author:    cleanStrings(doc.Metadata()["author"]),
		PageCount: doc.NumPage(),
		FilePath:  filePath,
		Hash:      hash,
	}

	// Fallback to filename if PDF metadata is empty
	if info.Title == "" || info.Title == "Untitled" {
		base := filepath.Base(filePath)
		info.Title = strings.TrimSuffix(base, filepath.Ext(base))
	}

	// Extract the cover
	img, err := doc.Image(0)
	if err != nil {
		// Return info even if cover extraction fails
		return info, nil, nil
	}

	var buf bytes.Buffer
	// Compress to JPEG to save memory and disk space
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})

	return info, buf.Bytes(), nil
}

// RenderPage is used later for the Reader mode to render specific pages.
func RenderPage(filePath string, pageNum int) (*bytes.Buffer, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	if pageNum >= doc.NumPage() || pageNum < 0 {
		return nil, nil
	}

	img, err := doc.Image(pageNum)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return &buf, nil
}
