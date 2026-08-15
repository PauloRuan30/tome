package pdf

import (
	"bytes"
	"image/jpeg"
	"path/filepath"
	"strings"

	"github.com/gen2brain/go-fitz"
)

type BookInfo struct {
	Title     string
	Author    string
	PageCount int
	FilePath  string
	Hash      string
}

func cleanStr(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(s, "\x00", ""))
}

func extractInfo(doc *fitz.Document, filePath, hash string) *BookInfo {
	meta := doc.Metadata()
	info := &BookInfo{
		Title:     cleanStr(meta["title"]),
		Author:    cleanStr(meta["author"]),
		PageCount: doc.NumPage(),
		FilePath:  filePath,
		Hash:      hash,
	}
	if info.Title == "" || info.Title == "Untitled" {
		base := filepath.Base(filePath)
		info.Title = strings.TrimSuffix(base, filepath.Ext(base))
	}
	return info
}

// Parse renders the cover at 72 DPI. That's plenty for a terminal
// thumbnail and ~10x faster than the 300 DPI default.
func Parse(filePath string, hash string) (*BookInfo, []byte, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer doc.Close()

	info := extractInfo(doc, filePath, hash)

	img, err := doc.ImageDPI(0, 72)
	if err != nil {
		return info, nil, nil
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80})
	return info, buf.Bytes(), nil
}

// ParseMetadataOnly skips image rendering entirely (used on cache hits).
func ParseMetadataOnly(filePath string, hash string) (*BookInfo, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	return extractInfo(doc, filePath, hash), nil
}

// RenderPage keeps a higher DPI for the actual Reader (Step 5).
func RenderPage(filePath string, pageNum int) (*bytes.Buffer, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	if pageNum >= doc.NumPage() || pageNum < 0 {
		return nil, nil
	}

	img, err := doc.ImageDPI(pageNum, 110)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90})
	return &buf, nil
}
