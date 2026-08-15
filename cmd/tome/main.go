package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/PauloRuan30/tome/internal/cache"
	"github.com/PauloRuan30/tome/internal/config"
	"github.com/PauloRuan30/tome/internal/pdf"
	"github.com/PauloRuan30/tome/internal/progress"
	"github.com/PauloRuan30/tome/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	version     = "0.1.0"
	dirFlag     = flag.String("dir", ".", "Directory to scan for PDFs")
	cleanFlag   = flag.Bool("clean-cache", false, "Clear the cover cache")
	kittyTest   = flag.Bool("kitty-test", false, "Paint one cover outside the TUI (Kitty debug)")
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
	tracker := progress.NewTracker()

	p := tea.NewProgram(
		tui.InitialModel(books, tracker),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if *kittyTest {
		files, _ := filepath.Glob(filepath.Join(*dirFlag, "*.pdf"))
		if len(files) == 0 {
			slog.Error("no pdfs found")
			os.Exit(1)
		}
		hash, _ := cache.HashFile(files[0])
		cover, err := cache.LoadCover(hash)
		if err != nil {
			_, cover, _ = pdf.Parse(files[0], hash)
		}

		// tiny 60x80 gradient JPEG => base64 < 4096 => single chunk
		tiny := image.NewRGBA(image.Rect(0, 0, 60, 80))
		for y := 0; y < 80; y++ {
			for x := 0; x < 60; x++ {
				tiny.Set(x, y, color.RGBA{uint8(x * 4), uint8(y * 3), 200, 255})
			}
		}
		var tinyBuf bytes.Buffer
		_ = jpeg.Encode(&tinyBuf, tiny, &jpeg.Options{Quality: 80})

		single := func(img []byte, row, col int, keys string) {
			b64 := base64.StdEncoding.EncodeToString(img)
			fmt.Printf("\x1b[%d;%dH", row, col)
			fmt.Printf("\x1b_G%s;q=2;%s\x1b\\", keys, b64)
		}
		chunked := func(img []byte, row, col int, keys string) {
			b64 := base64.StdEncoding.EncodeToString(img)
			fmt.Printf("\x1b[%d;%dH", row, col)
			const chunk = 4096
			for i := 0; i < len(b64); i += chunk {
				end := min(i+chunk, len(b64))
				m := 0
				if end < len(b64) {
					m = 1
				}
				if i == 0 {
					fmt.Printf("\x1b_G%s;q=2,m=%d;%s\x1b\\", keys, m, b64[i:end])
				} else {
					fmt.Printf("\x1b_Gm=%d;%s\x1b\\", m, b64[i:end])
				}
			}
		}

		cfg, _, _ := image.DecodeConfig(bytes.NewReader(cover))

		fmt.Println("quadrant test — T1 top-left, T2 top-right, T3 bottom-left, T4 bottom-right")
		single(tinyBuf.Bytes(), 3, 2, "a=T,f=100,s=60,v=80")                              // T1: small + s,v
		single(tinyBuf.Bytes(), 3, 40, "a=T,f=100,c=20,r=6")                              // T2: small + c,r
		chunked(cover, 12, 2, "a=T,f=100,c=40,r=28")                                      // T3: big + c,r (current way)
		chunked(cover, 12, 50, fmt.Sprintf("a=T,f=100,s=%d,v=%d", cfg.Width, cfg.Height)) // T4: big + s,v

		fmt.Println("sleeping 8s — tell me exactly which quadrants show an image")
		time.Sleep(8 * time.Second)
		pdf.ClearAllImages()
		return
	}

	if _, err := p.Run(); err != nil {
		pdf.ClearAllImages()
		slog.Error("TUI crashed", "err", err)
		os.Exit(1)
	}
	pdf.ClearAllImages()
}
