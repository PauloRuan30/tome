package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

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

	// --- THE WORKER POOL ---
	type result struct {
		book tui.Book
	}

	jobs := make(chan string, len(files))
	results := make(chan result, len(files))

	// Use all available CPU cores for maximum parsing speed
	workerCount := runtime.NumCPU()
	var wg sync.WaitGroup

	slog.Info("Starting worker pool to parse PDFs...", "workers", workerCount, "files", len(files))
	startTime := time.Now()

	for w := 0; w < workerCount; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				hash, err := cache.HashFile(path)
				if err != nil {
					continue
				}

				var cover []byte
				if cached, err := cache.LoadCover(hash); err == nil {
					// CACHE HIT: Load image instantly, parse only text metadata
					cover = cached
					info, _ := pdf.ParseMetadataOnly(path, hash)
					if info != nil {
						results <- result{book: tui.Book{Info: info, Cover: cover}}
					}
					continue
				}

				// CACHE MISS: Full parse and save
				info, c, err := pdf.Parse(path, hash)
				if err == nil && c != nil {
					cache.SaveCover(hash, c)
					cover = c
				}
				if info != nil {
					results <- result{book: tui.Book{Info: info, Cover: cover}}
				}
			}
		}()
	}

	// Feed the workers
	for _, f := range files {
		jobs <- f
	}
	close(jobs)

	// Wait for all parsing to finish
	wg.Wait()
	close(results)

	var books []tui.Book
	for r := range results {
		books = append(books, r.book)
	}

	slog.Info("Parsing complete", "duration", time.Since(startTime), "books", len(books))

	// Launch TUI
	p := tea.NewProgram(
		tui.InitialModel(books),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		slog.Error("TUI crashed", "err", err)
		os.Exit(1)
	}
}
