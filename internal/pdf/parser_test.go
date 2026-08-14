package pdf

import "testing"

func TestParseMissingFile(t *testing.T) {
	_, _, err := Parse("nonexistent.pdf", "fakehash")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}
