package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSearchModelTyping(t *testing.T) {
	m := NewSearchModel()

	// Type "hello"
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'h', 'e', 'l', 'l', 'o'}})
	if m.Query() != "hello" {
		t.Errorf("expected 'hello', got %q", m.Query())
	}

	// Backspace
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	if m.Query() != "hell" {
		t.Errorf("expected 'hell', got %q", m.Query())
	}

	// Reset
	m.Reset()
	if m.Query() != "" {
		t.Errorf("expected empty after reset, got %q", m.Query())
	}
}
