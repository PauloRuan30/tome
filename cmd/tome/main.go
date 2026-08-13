package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/yourusername/tome/internal/cache"
)

var (
	version     = "0.1.0"
	dirFlag     = flag.String("dir", ".", "Directory to scan for PDFs")
	cleanFlag   = flag.Bool("clean-cache", false, "Clear the cover cache")
	versionFlag = flag.Bool("version", false, "Print version")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("tome %s\n", version)
		return
	}

	if *cleanFlag {
		slog.Info("Cleaning cache...")
		os.RemoveAll(cache.CacheDir())
		return
	}

}
