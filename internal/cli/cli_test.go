package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir()) // ignore the developer's own map.yaml
	mapPath = ""
	root := newRootCmd()
	var out bytes.Buffer
	root.SetIn(strings.NewReader(stdin))
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

func TestCommands(t *testing.T) {
	tests := []struct {
		name  string
		stdin string
		args  []string
		want  string
	}{
		{"encode args", "", []string{"encode", "Hello", "World"}, "#3110 \\/\\/0|21|)\n"},
		{"encode alias and level", "", []string{"to", "-l", "basic", "Hello"}, "H3110\n"},
		{"encode stdin lines", "hi\nyo\n", []string{"encode"}, "#!\n`/0\n"},
		{"encode seeded random", "", []string{"encode", "-r", "--seed", "42", "Hello"}, "|-|&11[]\n"},
		{"decode", "", []string{"decode", "#3110"}, "hello\n"},
		{"decode stdin", "#!\n", []string{"from"}, "hi\n"},
		{"detect", "", []string{"detect", "h4<|<"}, "score    75% (mostly 1337)\ndecoded  hack\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := run(t, tt.stdin, tt.args...)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTableListsEveryLetter(t *testing.T) {
	got, err := run(t, "", "table")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"LETTER", "elite", `\/\/`, "|_|"} {
		if !strings.Contains(got, want) {
			t.Errorf("table output lacks %q", want)
		}
	}
}

func TestCustomMap(t *testing.T) {
	path := filepath.Join(t.TempDir(), "map.yaml")
	if err := os.WriteFile(path, []byte(`a: "@"`), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := run(t, "", "--map", path, "encode", "banana")
	if err != nil {
		t.Fatal(err)
	}
	if got != "8@|\\|@|\\|@\n" {
		t.Errorf("got %q", got)
	}
}

func TestErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"bad level":       {"encode", "-l", "mega", "x"},
		"missing map":     {"--map", "/nonexistent/map.yaml", "encode", "x"},
		"table with args": {"table", "x"},
	} {
		if _, err := run(t, "", args...); err == nil {
			t.Errorf("%s: expected an error", name)
		}
	}
}
