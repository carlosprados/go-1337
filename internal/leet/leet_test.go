package leet

import (
	"math/rand/v2"
	"strings"
	"testing"
)

func TestEncodeLevels(t *testing.T) {
	tests := []struct {
		level Level
		in    string
		want  string
	}{
		{Basic, "Hello World", "H3110 W0r1d"},
		{Advanced, "Hello World", "#3110 W0r1d"},
		{Elite, "Hello World", `#3110 \/\/0|21|)`},
		{Elite, "leet", "1337"},
		{Basic, "leet", "1337"},
		{Elite, "ñandú 42", "ñ4|\\||)ú 42"},
		{Elite, "", ""},
	}
	for _, tt := range tests {
		if got := Encode(tt.in, tt.level); got != tt.want {
			t.Errorf("Encode(%q, %v) = %q, want %q", tt.in, tt.level, got, tt.want)
		}
	}
}

func TestPrimaryRoundTrip(t *testing.T) {
	// Holds for text whose non-letters are not leet variants: "Hi!" comes back as "hii".
	const pangram = "The quick brown fox jumps over the lazy dog, twice."
	for _, level := range []Level{Basic, Advanced, Elite} {
		if got := Decode(Encode(pangram, level)); got != strings.ToLower(pangram) {
			t.Errorf("level %v: round trip = %q", level, got)
		}
	}
}

func TestDecode(t *testing.T) {
	tests := map[string]string{
		`\/\/`:    "w", // not "vv"
		`|_|`:     "u", // not "l" + "|"
		`|-|`:     "h",
		"1":       "l", // primary of l beats secondary of i
		"!":       "i",
		"#3110":   "hello",
		"PLAIN":   "plain",
		"(,)":     "q",
		"9 ^^ ^/": "g m n",
		"日本 1337": "日本 leet",
	}
	for in, want := range tests {
		if got := Decode(in); got != want {
			t.Errorf("Decode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDecodeIsDeterministic(t *testing.T) {
	in := Encode("we will win", Elite)
	first := Decode(in)
	for range 200 {
		if got := Decode(in); got != first {
			t.Fatalf("Decode(%q) changed between runs: %q vs %q", in, first, got)
		}
	}
}

func TestDefaultAlphabetInvariants(t *testing.T) {
	primaries := map[string]rune{}
	for _, e := range Default().Table() {
		if prev, dup := primaries[e.Variants[0]]; dup {
			t.Errorf("primary %q shared by %q and %q", e.Variants[0], prev, e.Letter)
		}
		primaries[e.Variants[0]] = e.Letter
		for _, v := range e.Variants {
			if got := Decode(v); got != string(e.Letter) && !(e.Letter == 'i' && v == "1") {
				t.Errorf("variant %q of %q decodes as %q", v, e.Letter, got)
			}
		}
	}
	if len(primaries) != 26 {
		t.Errorf("alphabet has %d letters, want 26", len(primaries))
	}
}

func TestRandomEncodingIsSeededAndValid(t *testing.T) {
	enc := func(seed uint64) string {
		return Default().Encode("aaaaaaaaaaaaaaaaaaaa", EncodeOptions{Level: Elite, Rand: rand.New(rand.NewPCG(seed, seed))})
	}
	if a, b := enc(1), enc(1); a != b {
		t.Error("same seed produced different output")
	}
	if got := Decode(enc(1)); got != strings.Repeat("a", 20) {
		t.Errorf("random a's decode as %q", got)
	}
}

func TestOverride(t *testing.T) {
	a, err := Default().Override(map[string][]string{"A": {"@"}, "e": {"3", "&"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := a.Encode("ape", EncodeOptions{Level: Elite}); got != "@|*3" {
		t.Errorf("Encode = %q", got)
	}
	if got := a.Decode("@|*&"); got != "ape" {
		t.Errorf("Decode = %q", got)
	}
	if Default().Encode("a", EncodeOptions{Level: Elite}) != "4" {
		t.Error("Override mutated the default alphabet")
	}

	for name, bad := range map[string]map[string][]string{
		"letter in variant": {"a": {"ah"}},
		"space in variant":  {"a": {"4 "}},
		"empty list":        {"a": {}},
		"empty variant":     {"a": {""}},
		"multi-letter key":  {"ab": {"4"}},
		"non-letter key":    {"1": {"!"}},
	} {
		if _, err := Default().Override(bad); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}

func TestParseOverrides(t *testing.T) {
	got, err := ParseOverrides([]byte("a: \"@\"\ne: [\"3\", \"&\"]\n"))
	if err != nil {
		t.Fatal(err)
	}
	if len(got["a"]) != 1 || got["a"][0] != "@" || len(got["e"]) != 2 {
		t.Errorf("ParseOverrides = %v", got)
	}
	if _, err := ParseOverrides([]byte("a: {x: 1}")); err == nil {
		t.Error("expected an error for a mapping value")
	}
}

func TestDetect(t *testing.T) {
	tests := []struct {
		in      string
		verdict string
		decoded string
	}{
		{"just normal text", "plain text", "just normal text"},
		{Encode("hack the planet", Elite), "full 1337", "hack the planet"},
		{"h4ck th3 p14n37", "mostly 1337", "hack the planet"},
		{"... ???", "no letters", "... ???"},
	}
	for _, tt := range tests {
		s := Default().Detect(tt.in)
		if s.Verdict() != tt.verdict || s.Decoded != tt.decoded {
			t.Errorf("Detect(%q) = %.2f %q %q", tt.in, s.Ratio, s.Verdict(), s.Decoded)
		}
	}
}

func TestParseLevel(t *testing.T) {
	for i, name := range Levels() {
		if l, err := ParseLevel(strings.ToUpper(name)); err != nil || l != Level(i) {
			t.Errorf("ParseLevel(%q) = %v, %v", name, l, err)
		}
	}
	if _, err := ParseLevel("mega"); err == nil {
		t.Error("expected an error")
	}
}
