// Package anim renders the "decrypting" effect: characters cycle through
// random leet glyphs and lock into place roughly left to right.
//
// Frames are a pure function of (target, progress, seed), so callers drive
// time however they like: a terminal loop, a Bubble Tea tick or
// requestAnimationFrame in the browser.
package anim

import (
	"time"
	"unicode"
)

const glyphs = `!#$%&*+<>?@[]{}|/\^~=_()0123456789`

// Duration is how long the effect lasts, the same in every front end.
const Duration = 2 * time.Second

const (
	// changes is how many times the scrambled glyphs change over a full
	// animation: about 20 per second.
	changes = 40
	// scrambleOnly is the opening share of the animation where nothing locks
	// yet, so the scramble reads before the text resolves.
	scrambleOnly = 0.25
)

// Cell is one rune of a frame.
type Cell struct {
	R      rune
	Locked bool // true once the rune shows its final value
}

// Cells returns the frame at progress (0 = fully scrambled, 1 = target).
// Whitespace is never scrambled, so line breaks and word shapes stay put.
func Cells(target string, progress float64, seed uint64) []Cell {
	runes := []rune(target)
	cells := make([]Cell, len(runes))
	bucket := uint64(progress * changes)
	for i, r := range runes {
		if progress >= 1 || unicode.IsSpace(r) || progress >= lockAt(i, len(runes), seed) {
			cells[i] = Cell{R: r, Locked: true}
			continue
		}
		h := mix(seed ^ uint64(i)*0x9e3779b97f4a7c15 ^ bucket*0xbf58476d1ce4e5b9)
		cells[i] = Cell{R: rune(glyphs[h%uint64(len(glyphs))])}
	}
	return cells
}

// Frame is Cells as a plain string.
func Frame(target string, progress float64, seed uint64) string {
	cells := Cells(target, progress, seed)
	out := make([]rune, len(cells))
	for i, c := range cells {
		out[i] = c.R
	}
	return string(out)
}

// lockAt is when rune i settles: after the scramble-only opening, a
// left-to-right sweep with up to 30% of jitter so it does not look mechanical.
func lockAt(i, n int, seed uint64) float64 {
	pos := float64(i) / float64(max(n, 1))
	jitter := float64(mix(seed^uint64(i)*0x94d049bb133111eb)%1000) / 1000
	return scrambleOnly + (1-scrambleOnly)*(0.7*pos+0.3*jitter)
}

// mix is splitmix64's finaliser: a cheap, well-distributed hash.
func mix(x uint64) uint64 {
	x ^= x >> 30
	x *= 0xbf58476d1ce4e5b9
	x ^= x >> 27
	x *= 0x94d049bb133111eb
	return x ^ x>>31
}
