package anim

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestFrameEndpoints(t *testing.T) {
	const target = "#4<|< 7#3\n|*14|\\|37 ñ"
	if got := Frame(target, 1, 7); got != target {
		t.Errorf("progress 1 = %q, want the target", got)
	}
	start := Frame(target, 0, 7)
	if utf8.RuneCountInString(start) != utf8.RuneCountInString(target) {
		t.Errorf("frame length changed: %q", start)
	}
	for i, r := range []rune(target) {
		if s := []rune(start)[i]; (r == ' ' || r == '\n') != (s == ' ' || s == '\n') {
			t.Errorf("whitespace not preserved at %d: %q", i, start)
		}
	}
}

func TestFrameIsDeterministicAndLocksMonotonically(t *testing.T) {
	target := strings.Repeat("hack the planet ", 4)
	if a, b := Frame(target, 0.4, 1), Frame(target, 0.4, 1); a != b {
		t.Fatal("same inputs gave different frames")
	}
	prev := 0
	for p := 0.0; p <= 1.0; p += 0.05 {
		locked := 0
		for _, c := range Cells(target, p, 1) {
			if c.Locked {
				locked++
			}
		}
		if locked < prev {
			t.Fatalf("locked cells went down from %d to %d at %.2f", prev, locked, p)
		}
		prev = locked
	}
}

func TestFrameScramblesBeforeTheEnd(t *testing.T) {
	target := "abcdefghijklmnopqrstuvwxyz"
	if Frame(target, 0, 3) == target {
		t.Error("progress 0 shows the target unchanged")
	}
	if Frame(target, 0.1, 3) == Frame(target, 0.5, 3) {
		t.Error("frames do not change over time")
	}
}

func TestNothingLocksDuringTheOpening(t *testing.T) {
	for _, c := range Cells("hack the planet", scrambleOnly-0.01, 9) {
		if c.Locked && c.R != ' ' {
			t.Fatalf("%q locked before the scramble-only phase ended", c.R)
		}
	}
}
