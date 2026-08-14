package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/PauloRuan30/tome/internal/cache"
	"github.com/PauloRuan30/tome/internal/config"
)

var (
	version     = "0.1.0"
	cleanFlag   = flag.Bool("clean-cache", false, "Clear the cover cache")
	versionFlag = flag.Bool("version", false, "Print version")
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("tome %s\n", version)
		return
	}

	// Test Config Loading
	cfg := config.Load()
	slog.Info("Configuration loaded", "settings", cfg)

	if *cleanFlag {
		slog.Info("Cleaning cache directory...", "path", cache.CacheDir())
		os.RemoveAll(cache.CacheDir())
		return
	}

	slog.Info("It worked!")
}
