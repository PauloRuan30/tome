package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// CacheDir returns the standard OS cache directory for tome.
func CacheDir() string {
	dir, _ := os.UserCacheDir()
	return filepath.Join(dir, "tome", "covers")
}

// HashFile reads the first 64KB of a file and returns a SHA-256 hex string.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.CopyN(h, f, 64*1024); err != nil && err != io.EOF {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// SaveCover writes image data to the cache directory.
func SaveCover(hash string, data []byte) error {
	dir := CacheDir()
	os.MkdirAll(dir, 0755)
	return os.WriteFile(filepath.Join(dir, hash+".jpg"), data, 0644)
}

// attempts to read a cached cover image.
func LoadCover(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(CacheDir(), hash+".jpg"))
}
