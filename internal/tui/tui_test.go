package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/carlosprados/go-1337/internal/leet"
	tea "github.com/charmbracelet/bubbletea"
)

func press(m Model, msgs ...tea.Msg) Model {
	for _, msg := range msgs {
		next, _ := m.Update(msg)
		m = next.(Model)
	}
	return m
}

func typed(s string) tea.Msg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestEditorConvertsAsYouType(t *testing.T) {
	var copied string
	m := New(leet.Default(), func(s string) error { copied = s; return nil })
	m = press(m, tea.WindowSizeMsg{Width: 120, Height: 30}, typed("hello"))

	if got := m.Output(); got != "#3110" {
		t.Errorf("elite output = %q", got)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlL}) // elite -> basic
	if got := m.Output(); got != "h3110" {
		t.Errorf("basic output = %q", got)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyTab})
	if got := m.Output(); got != "hello" {
		t.Errorf("decode output = %q", got)
	}

	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlY})
	if copied != "hello" {
		t.Errorf("copied %q", copied)
	}
	if view := m.View(); !strings.Contains(view, "copied to clipboard") {
		t.Error("view does not show the copy status")
	}
}

func TestRandomIsStableUntilReshuffle(t *testing.T) {
	m := New(leet.Default(), nil)
	m = press(m, typed("aaaaaaaaaaaa"), tea.KeyMsg{Type: tea.KeyCtrlR})
	first := m.Output()
	if m.Output() != first {
		t.Fatal("random output changed without input")
	}
	m = press(m, tea.KeyMsg{Type: tea.KeyCtrlS})
	if m.Output() == first {
		t.Error("reshuffle kept the same output") // 3^12 combinations: a collision is not realistic
	}
}

func TestTogglesPlayTheAnimation(t *testing.T) {
	m := New(leet.Default(), nil)
	m = press(m, tea.WindowSizeMsg{Width: 160, Height: 30}, typed("the quick brown fox jumps over the lazy dog"))
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlL})
	m = next.(Model)
	if cmd == nil || m.animProgress() >= 1 {
		t.Fatal("ctrl+l did not start the animation")
	}
	if strings.Contains(m.View(), m.Output()) {
		t.Error("view shows the final output while animating")
	}

	m.animStart = time.Now().Add(-time.Minute) // let it finish
	next, cmd = m.Update(animTickMsg{})
	m = next.(Model)
	if cmd != nil || m.animProgress() != 1 {
		t.Error("animation did not stop after its duration")
	}
	if !strings.Contains(m.View(), m.Output()) {
		t.Error("view lacks the final output after the animation")
	}
}
