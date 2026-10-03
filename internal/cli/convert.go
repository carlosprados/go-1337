package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/carlosprados/go-1337/internal/leet"
	"github.com/spf13/cobra"
)

const inputHelp = `
With arguments, they are joined with spaces and converted as one line.
Without arguments, every line read from standard input is converted.`

func newEncodeCmd() *cobra.Command {
	var (
		level  string
		random bool
		seed   uint64
		toClip bool
	)
	cmd := &cobra.Command{
		Use:     "encode [text...]",
		Aliases: []string{"to", "to1337"},
		Short:   "Convert plain text to 1337",
		Long: `Convert plain text to 1337 speak.

--level picks how much gets converted:
  basic     a e l o s t        (h3llo -> #3110 at elite, h3110 at basic)
  advanced  basic + b c g h i k z
  elite     the whole alphabet (default)

By default every letter uses its primary variant, so the output decodes back
exactly. --random picks a random variant per letter instead (see "leet table").
` + inputHelp,
		Example: `  leet encode Hello World
  leet encode -l basic "leet speak"
  leet encode --random --seed 42 "same output every time"
  leet encode -c "straight to the clipboard"
  leet encode < notes.txt > notes.1337`,
		RunE: func(cmd *cobra.Command, args []string) error {
			lvl, err := leet.ParseLevel(level)
			if err != nil {
				return err
			}
			a, err := alphabet()
			if err != nil {
				return err
			}
			opts := leet.EncodeOptions{Level: lvl}
			if random {
				if seed == 0 {
					seed = uint64(time.Now().UnixNano())
				}
				opts.Rand = rand.New(rand.NewPCG(seed, seed))
			}
			return convert(cmd, args, toClip, func(s string) string { return a.Encode(s, opts) })
		},
	}
	f := cmd.Flags()
	f.StringVarP(&level, "level", "l", leet.Elite.String(), "how much to convert: "+strings.Join(leet.Levels(), ", "))
	f.BoolVarP(&random, "random", "r", false, "pick a random variant for each letter")
	f.Uint64Var(&seed, "seed", 0, "seed for --random, for reproducible output (0 = time-based)")
	f.BoolVarP(&toClip, "copy", "c", false, "also copy the result to the clipboard")
	_ = cmd.RegisterFlagCompletionFunc("level", cobra.FixedCompletions(leet.Levels(), cobra.ShellCompDirectiveNoFileComp))
	return cmd
}

func newDecodeCmd() *cobra.Command {
	var toClip bool
	cmd := &cobra.Command{
		Use:     "decode [text...]",
		Aliases: []string{"from", "from1337"},
		Short:   "Convert 1337 back to plain text",
		Long: `Convert 1337 speak back to lowercase plain text.

Every variant of every letter is recognised, whatever level or randomness
produced it. The longest sequence wins, and on a tie the primary variant
wins: "1" decodes as "l" (it is also a secondary "i").

Case is not recovered, and digits in the input are read as leet.
Quote the argument in your shell: many leet sequences use \ | < > * $.
` + inputHelp,
		Example: `  leet decode '#3110 \/\/0|21|)'
  leet encode "Hello" | leet decode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := alphabet()
			if err != nil {
				return err
			}
			return convert(cmd, args, toClip, a.Decode)
		},
	}
	cmd.Flags().BoolVarP(&toClip, "copy", "c", false, "also copy the result to the clipboard")
	return cmd
}

func convert(cmd *cobra.Command, args []string, toClip bool, fn func(string) string) error {
	var copied bytes.Buffer
	out := cmd.OutOrStdout()
	if toClip {
		out = io.MultiWriter(out, &copied)
	}
	if len(args) > 0 {
		if _, err := fmt.Fprintln(out, fn(strings.Join(args, " "))); err != nil {
			return err
		}
	} else {
		if isTerminal(cmd.InOrStdin()) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Type text and press Enter (Ctrl-D to finish):")
		}
		if err := convertLines(cmd.InOrStdin(), out, fn); err != nil {
			return err
		}
	}
	if toClip {
		return copyToClipboard(strings.TrimSuffix(copied.String(), "\n"))
	}
	return nil
}

func convertLines(in io.Reader, out io.Writer, fn func(string) string) error {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if _, err := fmt.Fprintln(out, fn(scanner.Text())); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// readInput returns the joined args or, without args, the whole of stdin.
func readInput(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}
	if isTerminal(cmd.InOrStdin()) {
		fmt.Fprintln(cmd.ErrOrStderr(), "Type text (Ctrl-D to finish):")
	}
	data, err := io.ReadAll(cmd.InOrStdin())
	return string(data), err
}
