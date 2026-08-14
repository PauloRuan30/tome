package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/PauloRuan30/tome/internal/cache"
	"github.com/PauloRuan30/tome/internal/config"
	"github.com/PauloRuan30/tome/internal/pdf"
)

var (
	version     = "0.1.0"
	dirFlag     = flag.String("dir", ".", "Directory to scan for PDFs")
	cleanFlag   = flag.Bool("clean-cache", false, "Clear the cover cache")
	versionFlag = flag.Bool("version", false, "Print version")
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("tome %s\n", version)
		return
	}

	cfg := config.Load()
	_ = cfg // Will be used in later steps

	if *cleanFlag {
		slog.Info("Cleaning cache directory...", "path", cache.CacheDir())
		os.RemoveAll(cache.CacheDir())
		return
	}

	// 1. Scan directory for PDFs
	files, err := filepath.Glob(filepath.Join(*dirFlag, "*.pdf"))
	if err != nil {
		slog.Error("Failed to read directory", "err", err)
		os.Exit(1)
	}

	slog.Info("Scanning for PDFs...", "dir", *dirFlag, "count", len(files))

	// 2. Process each PDF
	for _, path := range files {
		hash, err := cache.HashFile(path)
		if err != nil {
			slog.Warn("Failed to hash file", "path", path, "err", err)
			continue
		}

		// Check cache first
		if _, err := cache.LoadCover(hash); err == nil {
			slog.Info("Loaded from cache", "file", filepath.Base(path))
			continue
		}

		// Parse and cache
		info, cover, err := pdf.Parse(path, hash)
		if err != nil {
			slog.Warn("Failed to parse PDF", "path", path, "err", err)
			continue
		}

		if cover != nil {
			cache.SaveCover(hash, cover)
			slog.Info("Parsed and cached new book", "title", info.Title, "pages", info.PageCount)
		} else {
			slog.Warn("No cover extracted", "title", info.Title)
		}
	}

	slog.Info("Scan complete.")
}
