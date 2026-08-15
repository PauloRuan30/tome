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
	"github.com/PauloRuan30/tome/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
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

	_ = config.Load()

	if *cleanFlag {
		slog.Info("Cleaning cache directory...", "path", cache.CacheDir())
		os.RemoveAll(cache.CacheDir())
		return
	}

	files, err := filepath.Glob(filepath.Join(*dirFlag, "*.pdf"))
	if err != nil {
		slog.Error("Failed to read directory", "err", err)
		os.Exit(1)
	}

	var books []tui.Book
	for _, path := range files {
		hash, err := cache.HashFile(path)
		if err != nil {
			continue
		}

		var cover []byte
		if cached, err := cache.LoadCover(hash); err == nil {
			cover = cached
		} else {
			_, cover, _ = pdf.Parse(path, hash)
			if cover != nil {
				cache.SaveCover(hash, cover)
			}
		}

		// We need the info for the UI, so we parse it quickly
		info, _, _ := pdf.Parse(path, hash)
		if info != nil {
			books = append(books, tui.Book{Info: info, Cover: cover})
		}
	}

	// Launch the TUI using the 'tea' alias
	p := tea.NewProgram(
		tui.InitialModel(books),
		tea.WithAltScreen(),       // Use the alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse tracking
	)

	if _, err := p.Run(); err != nil {
		slog.Error("TUI crashed", "err", err)
		os.Exit(1)
	}
}
