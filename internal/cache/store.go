package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

func CacheDir() string {
	dir, _ := os.UserCacheDir()
	return filepath.Join(dir, "tome", "covers")
}

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

func SaveCover(hash string, data []byte) error {
	dir := CacheDir()
	os.MkdirAll(dir, 0755)
	return os.WriteFile(filepath.Join(dir, hash+".png"), data, 0644) // Changed to .png
}

func LoadCover(hash string) ([]byte, error) {
	return os.ReadFile(filepath.Join(CacheDir(), hash+".png")) // Changed to .png
}
