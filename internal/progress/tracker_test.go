package progress

import (
	"path/filepath"
	"testing"
	"time"
)

func TestTrackerRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "progress.json")

	tr := NewTrackerWithPath(path)
	tr.Update("abc", Progress{LastPage: 42, Total: 300, LastOpen: time.Now()})

	loaded := NewTrackerWithPath(path)
	got := loaded.Get("abc")
	if got.LastPage != 42 || got.Total != 300 {
		t.Fatalf("roundtrip failed: %+v", got)
	}
	if loaded.Get("missing").LastPage != 0 {
		t.Fatal("missing key should be zero value")
	}
}
