package cache

import (
	"os"
	"testing"
)

func TestHashFile(t *testing.T) {
	// Create a temporary file with known content
	tmpFile, err := os.CreateTemp("", "test_*.pdf")
	if err != nil {
		t.Fatal("Failed to create temp file:", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write More than 64KB of data to the file
	content := make([]byte, 128*1024) // 128KB of data
	for i := range content {
		content[i] = 'A' + byte(i%26) // Fill with A-Z
	}

	tmpFile.Write(content)
	tmpFile.Close()

	// Calculate the hash using the function
	hash, err := HashFile(tmpFile.Name())
	if err != nil {
		t.Fatal("HashFile failed:", err)
	}

	// Sha-256 hash of the first 64KB of the content
	if len(hash) != 64 {
		t.Errorf("Expected hash length of 64, got %d", len(hash))
	}
}
