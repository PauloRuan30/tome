package pdf

import (
	"bytes"
	"image/png" // Switched to PNG
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
	png.Encode(&buf, img) // Encode as PNG!
	return info, buf.Bytes(), nil
}

func ParseMetadataOnly(filePath string, hash string) (*BookInfo, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()
	return extractInfo(doc, filePath, hash), nil
}

func RenderPage(filePath string, pageNum int) (*bytes.Buffer, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	if pageNum >= doc.NumPage() || pageNum < 0 {
		return nil, nil
	}

	img, err := doc.ImageDPI(pageNum, 150)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	png.Encode(&buf, img) // Encode as PNG!
	return &buf, nil
}

func PageText(filePath string, pageNum int) (string, error) {
	doc, err := fitz.New(filePath)
	if err != nil {
		return "", err
	}
	defer doc.Close()

	if pageNum < 0 || pageNum >= doc.NumPage() {
		return "", nil
	}
	return doc.Text(pageNum)
}
