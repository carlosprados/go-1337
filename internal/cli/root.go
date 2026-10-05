// Package cli implements the leet command-line interface.
package cli

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/atotto/clipboard"
	"github.com/carlosprados/go-1337/internal/leet"
	"github.com/carlosprados/go-1337/internal/tui"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// version is overridden at build time with -ldflags "-X github.com/carlosprados/go-1337/internal/cli.version=v1.2.3".
var version = "dev"

var mapPath string

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "leet",
		Short: "Convert text to and from 1337 (leet) speak",
		Long: `leet translates text into 1337 (leet) speak and back.

Text is taken from the command arguments or, when none are given, from
standard input line by line, so it works both interactively and in pipes.
Only the converted text is written to standard output.

Run it with no arguments in a terminal to open the live editor.

Custom alphabet: a YAML file mapping letters to one or more variants,
read from --map or, if it exists, from ` + defaultMapPathHint() + `:

  a: "@"
  e: ["3", "&"]`,
		Example: `  leet encode "Hello World"
  leet encode --level basic --random "Hello World"
  echo "Hello World" | leet encode | leet decode
  leet detect 'h4ck th3 p14n37'
  leet decode --animate '|*455\/\/0|2|) 4<<3|*73|)'
  leet share "meet me at the usual place"
  leet table`,
		Version:      version,
		SilenceUsage: true,
		Args:         cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if !isTerminal(os.Stdin) || !isTerminal(os.Stdout) {
				return cmd.Help()
			}
			return runTUI(cmd)
		},
	}
	root.PersistentFlags().StringVarP(&mapPath, "map", "m", "", "YAML file with a custom alphabet (default "+defaultMapPathHint()+" if present)")
	root.AddCommand(newEncodeCmd(), newDecodeCmd(), newDetectCmd(), newTableCmd(), newTUICmd(), newShareCmd())
	return root
}

// Execute runs the root command and exits with a non-zero status on failure.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func defaultMapPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "leet", "map.yaml")
}

func defaultMapPathHint() string {
	if p := defaultMapPath(); p != "" {
		return p
	}
	return "$XDG_CONFIG_HOME/leet/map.yaml"
}

// alphabet resolves the alphabet: an explicit --map must exist, the default
// config file is optional.
func alphabet() (*leet.Alphabet, error) {
	if mapPath != "" {
		return leet.LoadAlphabet(mapPath)
	}
	p := defaultMapPath()
	if p == "" {
		return leet.Default(), nil
	}
	a, err := leet.LoadAlphabet(p)
	if errors.Is(err, fs.ErrNotExist) {
		return leet.Default(), nil
	}
	return a, err
}

func newTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:     "tui",
		Aliases: []string{"live", "play"},
		Short:   "Open the live editor: type and watch the conversion as you go",
		Long: `Open a full-screen editor that converts as you type.

Keys: tab switches encode/decode, ctrl+l cycles the level, ctrl+r toggles
random variants (ctrl+s reshuffles), ctrl+y copies the output, esc quits.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error { return runTUI(cmd) },
	}
}

func runTUI(_ *cobra.Command) error {
	a, err := alphabet()
	if err != nil {
		return err
	}
	return tui.Run(a, copyToClipboard)
}

// isTerminal reports whether r is an interactive terminal. A Stat-based check
// is not enough: /dev/null is a character device too.
func isTerminal(r any) bool {
	f, ok := r.(*os.File)
	return ok && term.IsTerminal(int(f.Fd()))
}

func copyToClipboard(text string) error { return clipboard.WriteAll(text) }
