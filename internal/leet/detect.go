package leet

import "unicode"

// Score summarises how much of a text is written in leet.
type Score struct {
	Leet    int     // characters consumed by leet tokens, counted as letters
	Plain   int     // plain letters
	Ratio   float64 // Leet / (Leet + Plain); 0 when there are no letters at all
	Decoded string
}

// Verdict is a human label for the ratio.
func (s Score) Verdict() string {
	switch {
	case s.Leet+s.Plain == 0:
		return "no letters"
	case s.Ratio >= 0.85:
		return "full 1337"
	case s.Ratio >= 0.35:
		return "mostly 1337"
	case s.Ratio >= 0.1:
		return "a bit 1337"
	default:
		return "plain text"
	}
}

// Detect measures how leet a text is. Digits in ordinary prose count as leet
// too, so a text full of numbers scores higher than it should.
func (a *Alphabet) Detect(text string) Score {
	var s Score
	a.scan(text, func(t *token, raw string) {
		switch {
		case t != nil:
			s.Leet++
		case unicode.IsLetter([]rune(raw)[0]):
			s.Plain++
		}
	})
	if total := s.Leet + s.Plain; total > 0 {
		s.Ratio = float64(s.Leet) / float64(total)
	}
	s.Decoded = a.Decode(text)
	return s
}
