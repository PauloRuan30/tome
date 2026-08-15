package progress

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type Progress struct {
	LastPage int       `json:"last_page"`
	Total    int       `json:"total"`
	LastOpen time.Time `json:"last_open"`
}

type Tracker struct {
	Path string
	Data map[string]Progress
}

func NewTracker() *Tracker {
	dir, _ := os.UserConfigDir()
	return NewTrackerWithPath(filepath.Join(dir, "tome", "progress.json"))
}

func NewTrackerWithPath(path string) *Tracker {
	t := &Tracker{Path: path, Data: map[string]Progress{}}
	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, &t.Data)
	}
	return t
}

func (t *Tracker) Get(hash string) Progress { return t.Data[hash] }

func (t *Tracker) Update(hash string, p Progress) {
	t.Data[hash] = p
	_ = t.Save()
}

func (t *Tracker) Save() error {
	if err := os.MkdirAll(filepath.Dir(t.Path), 0755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(t.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(t.Path, b, 0644)
}
