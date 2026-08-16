package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/PauloRuan30/tome/internal/cache"
	"github.com/PauloRuan30/tome/internal/config"
	"github.com/PauloRuan30/tome/internal/pdf"
	"github.com/PauloRuan30/tome/internal/progress"
	"github.com/PauloRuan30/tome/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "0.1.0"

func main() {
	dirFlag := flag.String("dir", ".", "Directory to scan for PDFs")
	cleanFlag := flag.Bool("clean-cache", false, "Clear the cover cache")
	versionFlag := flag.Bool("version", false, "Print version")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("tome %s\n", version)
		return
	}

	// Production logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
	slog.SetDefault(logger)

	if *cleanFlag {
		fmt.Println("Cleaning cache...")
		os.RemoveAll(cache.CacheDir())
		return
	}

	_ = config.Load()
	books := loadLibrary(*dirFlag)
	tracker := progress.NewTracker()

	p := tea.NewProgram(
		tui.InitialModel(books, tracker),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		pdf.ClearAllImages()
		fmt.Fprintf(os.Stderr, "Fatal error: %v\n", err)
		os.Exit(1)
	}

	// Ensure terminal is clean on exit
	pdf.ClearAllImages()
}

// loadLibrary scans the directory and parses PDFs concurrently using a worker pool.
func loadLibrary(dir string) []tui.Book {
	files, err := filepath.Glob(filepath.Join(dir, "*.pdf"))
	if err != nil || len(files) == 0 {
		return nil
	}

	type result struct{ book tui.Book }
	jobs := make(chan string, len(files))
	results := make(chan result, len(files))

	workers := runtime.NumCPU()
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				hash, err := cache.HashFile(path)
				if err != nil {
					continue
				}

				if cached, err := cache.LoadCover(hash); err == nil {
					info, _ := pdf.ParseMetadataOnly(path, hash)
					if info != nil {
						results <- result{book: tui.Book{Info: info, Cover: cached}}
					}
					continue
				}

				// Slow path: Cache miss (parse and save)
				info, cover, err := pdf.Parse(path, hash)
				if err == nil && cover != nil {
					_ = cache.SaveCover(hash, cover)
				}
				if info != nil {
					results <- result{book: tui.Book{Info: info, Cover: cover}}
				}
			}
		}()
	}

	for _, f := range files {
		jobs <- f
	}
	close(jobs)
	wg.Wait()
	close(results)

	var books []tui.Book
	for r := range results {
		books = append(books, r.book)
	}
	return books
}
