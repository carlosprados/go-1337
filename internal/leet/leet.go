// Package leet converts text to and from 1337 (leet) speak.
//
// Every letter has an ordered list of variants; the first one is its primary
// form, used for deterministic encoding. Decoding scans the input with
// longest-match-first over every variant of every letter, so it is
// deterministic and independent of the level used to encode.
package leet

import (
	"fmt"
	"maps"
	"math/rand/v2"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Level selects which letters get converted when encoding.
type Level int

const (
	Basic    Level = iota // a e l o s t: the classic "1337" letters
	Advanced              // basic plus b c g h i k z
	Elite                 // the whole alphabet
)

var levelNames = []string{"basic", "advanced", "elite"}

func (l Level) String() string {
	if l < Basic || l > Elite {
		return fmt.Sprintf("Level(%d)", int(l))
	}
	return levelNames[l]
}

// Levels returns every level name, lowest first.
func Levels() []string { return append([]string(nil), levelNames...) }

// ParseLevel parses a level name, case-insensitively.
func ParseLevel(s string) (Level, error) {
	for i, name := range levelNames {
		if strings.EqualFold(s, name) {
			return Level(i), nil
		}
	}
	return 0, fmt.Errorf("unknown level %q (valid: %s)", s, strings.Join(levelNames, ", "))
}

var defaultVariants = map[rune][]string{
	'a': {"4", "@", "/-\\"},
	'b': {"8", "|3"},
	'c': {"<", "(", "{"},
	'd': {"|)", "[)", "|>"},
	'e': {"3", "&", "[-"},
	'f': {"|=", "/="},
	'g': {"6", "9", "(_+"},
	'h': {"#", "|-|", "}{"},
	'i': {"!", "1"},
	'j': {"]", "_|"},
	'k': {"|<", "|{"},
	'l': {"1", "|_"},
	'm': {"/\\/\\", "|\\/|", "^^"},
	'n': {"|\\|", "^/"},
	'o': {"0", "()", "[]"},
	'p': {"|*", "|^"},
	'q': {"(,)", "0_"},
	'r': {"|2", "|?"},
	's': {"5", "$"},
	't': {"7", "+"},
	'u': {"|_|", "(_)"},
	'v': {"\\/"},
	'w': {"\\/\\/", "\\^/"},
	'x': {"><", ")("},
	'y': {"`/", "'/"},
	'z': {"2"},
}

// minLevel is the lowest level at which each letter is converted.
var minLevel = func() map[rune]Level {
	m := make(map[rune]Level, 26)
	for r := 'a'; r <= 'z'; r++ {
		m[r] = Elite
	}
	for _, r := range "aelost" {
		m[r] = Basic
	}
	for _, r := range "bcghikz" {
		m[r] = Advanced
	}
	return m
}()

// Alphabet is an immutable letter-to-variants table plus its decoding index.
type Alphabet struct {
	variants map[rune][]string
	tokens   []token
}

type token struct {
	text     string
	letter   rune
	priority int // index in the letter's variant list; 0 is the primary
}

var defaultAlphabet = mustNew(defaultVariants)

// Default returns the built-in alphabet.
func Default() *Alphabet { return defaultAlphabet }

func mustNew(variants map[rune][]string) *Alphabet {
	a, err := newAlphabet(variants)
	if err != nil {
		panic(err)
	}
	return a
}

func newAlphabet(variants map[rune][]string) (*Alphabet, error) {
	a := &Alphabet{variants: make(map[rune][]string, len(variants))}
	for letter, vs := range variants {
		if err := validate(letter, vs); err != nil {
			return nil, err
		}
		a.variants[letter] = append([]string(nil), vs...)
		for i, v := range vs {
			a.tokens = append(a.tokens, token{text: v, letter: letter, priority: i})
		}
	}
	// Longest first; on equal text the primary variant wins, then the lower letter.
	sort.Slice(a.tokens, func(i, j int) bool {
		ti, tj := a.tokens[i], a.tokens[j]
		if len(ti.text) != len(tj.text) {
			return len(ti.text) > len(tj.text)
		}
		if ti.text != tj.text {
			return ti.text < tj.text
		}
		if ti.priority != tj.priority {
			return ti.priority < tj.priority
		}
		return ti.letter < tj.letter
	})
	return a, nil
}

func validate(letter rune, vs []string) error {
	if letter < 'a' || letter > 'z' {
		return fmt.Errorf("key %q: must be a single letter a-z", letter)
	}
	if len(vs) == 0 {
		return fmt.Errorf("letter %q: needs at least one variant", letter)
	}
	for _, v := range vs {
		if v == "" {
			return fmt.Errorf("letter %q: empty variant", letter)
		}
		for _, r := range v {
			if unicode.IsLetter(r) || unicode.IsSpace(r) {
				return fmt.Errorf("letter %q: variant %q must not contain letters or spaces", letter, v)
			}
		}
	}
	return nil
}

// Override returns a new alphabet where the given letters use the given
// variants. Keys are single letters, case-insensitive.
func (a *Alphabet) Override(overrides map[string][]string) (*Alphabet, error) {
	merged := make(map[rune][]string, len(a.variants))
	maps.Copy(merged, a.variants)
	for k, vs := range overrides {
		r, size := utf8.DecodeRuneInString(strings.ToLower(k))
		if size != len(k) || len(k) == 0 {
			return nil, fmt.Errorf("key %q: must be a single letter a-z", k)
		}
		merged[r] = vs
	}
	return newAlphabet(merged)
}

// Entry describes one letter of an alphabet.
type Entry struct {
	Letter   rune
	Variants []string
	Level    Level // lowest level that converts this letter
}

// Table returns every letter with its variants, sorted alphabetically.
func (a *Alphabet) Table() []Entry {
	entries := make([]Entry, 0, len(a.variants))
	for r, vs := range a.variants {
		entries = append(entries, Entry{Letter: r, Variants: append([]string(nil), vs...), Level: minLevel[r]})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Letter < entries[j].Letter })
	return entries
}

// EncodeOptions tunes Encode.
type EncodeOptions struct {
	Level Level
	Rand  *rand.Rand // when set, a random variant is picked per letter
}

// Encode converts plain text to leet speak. Letters above the level and
// anything that is not a letter pass through unchanged, case preserved.
func (a *Alphabet) Encode(text string, opts EncodeOptions) string {
	var b strings.Builder
	for _, ch := range text {
		lower := unicode.ToLower(ch)
		vs, ok := a.variants[lower]
		if !ok || minLevel[lower] > opts.Level {
			b.WriteRune(ch)
			continue
		}
		v := vs[0]
		if opts.Rand != nil {
			v = vs[opts.Rand.IntN(len(vs))]
		}
		b.WriteString(v)
	}
	return b.String()
}

// Decode converts leet speak back to lowercase plain text.
func (a *Alphabet) Decode(text string) string {
	var b strings.Builder
	a.scan(text, func(t *token, raw string) {
		if t != nil {
			b.WriteRune(t.letter)
		} else {
			b.WriteString(strings.ToLower(raw))
		}
	})
	return b.String()
}

// scan walks text with longest-match-first, calling fn with the matched
// token, or with nil and the single unmatched rune.
func (a *Alphabet) scan(text string, fn func(t *token, raw string)) {
	for i := 0; i < len(text); {
		if t := a.match(text[i:]); t != nil {
			fn(t, t.text)
			i += len(t.text)
			continue
		}
		_, size := utf8.DecodeRuneInString(text[i:])
		fn(nil, text[i:i+size])
		i += size
	}
}

func (a *Alphabet) match(s string) *token {
	for i := range a.tokens {
		if strings.HasPrefix(s, a.tokens[i].text) {
			return &a.tokens[i]
		}
	}
	return nil
}

// Encode converts text with the default alphabet at the given level.
func Encode(text string, level Level) string {
	return defaultAlphabet.Encode(text, EncodeOptions{Level: level})
}

// Decode converts text back with the default alphabet.
func Decode(text string) string { return defaultAlphabet.Decode(text) }
