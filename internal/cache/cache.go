package cache

import (
	"log/slog"
	"os"
	"path/filepath"
)

// CacheDir returns the directory where rendered PDF covers are stored,
// creating it if it does not exist.
func CacheDir() string {
	base, err := os.UserCacheDir()
	if err != nil {
		slog.Warn("Could not find user cache dir", "err", err)
		base = os.TempDir()
	}

	dir := filepath.Join(base, "tome", "covers")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Warn("Could not create cache dir", "err", err, "dir", dir)
	}

	return dir
}
