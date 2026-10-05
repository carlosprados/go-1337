package cli

import (
	"bufio"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/carlosprados/go-1337/internal/anim"
	"github.com/carlosprados/go-1337/internal/leet"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/term"
)

const inputHelp = `
With arguments, they are joined with spaces and converted as one line.
Without arguments, every line read from standard input is converted.`

// encodeFlags are the flags shared by every command that encodes.
type encodeFlags struct {
	level  string
	random bool
	seed   uint64
}

func (f *encodeFlags) register(cmd *cobra.Command) {
	fs := cmd.Flags()
	fs.StringVarP(&f.level, "level", "l", leet.Elite.String(), "how much to convert: "+strings.Join(leet.Levels(), ", "))
	fs.BoolVarP(&f.random, "random", "r", false, "pick a random variant for each letter")
	fs.Uint64Var(&f.seed, "seed", 0, "seed for --random, for reproducible output (0 = time-based)")
	_ = cmd.RegisterFlagCompletionFunc("level", cobra.FixedCompletions(leet.Levels(), cobra.ShellCompDirectiveNoFileComp))
}

func (f *encodeFlags) options() (leet.EncodeOptions, error) {
	lvl, err := leet.ParseLevel(f.level)
	if err != nil {
		return leet.EncodeOptions{}, err
	}
	opts := leet.EncodeOptions{Level: lvl}
	if f.random {
		seed := f.seed
		if seed == 0 {
			seed = uint64(time.Now().UnixNano())
		}
		opts.Rand = rand.New(rand.NewPCG(seed, seed))
	}
	return opts, nil
}

// outputFlags control how converted lines are delivered.
type outputFlags struct {
	copy    bool
	animate bool
}

func (o *outputFlags) register(fs *pflag.FlagSet) {
	fs.BoolVarP(&o.copy, "copy", "c", false, "also copy the result to the clipboard")
	fs.BoolVarP(&o.animate, "animate", "a", false, "reveal the result with a decrypting animation (terminal only)")
}

func newEncodeCmd() *cobra.Command {
	var (
		ef  encodeFlags
		out outputFlags
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

--animate reveals each line hacker-movie style. It only plays when the output
is a terminal, so pipes and redirections always get plain text.
` + inputHelp,
		Example: `  leet encode Hello World
  leet encode -l basic "leet speak"
  leet encode --random --seed 42 "same output every time"
  leet encode --animate "access granted"
  leet encode -c "straight to the clipboard"
  leet encode < notes.txt > notes.1337`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := ef.options()
			if err != nil {
				return err
			}
			a, err := alphabet()
			if err != nil {
				return err
			}
			return convert(cmd, args, out, func(s string) string { return a.Encode(s, opts) })
		},
	}
	ef.register(cmd)
	out.register(cmd.Flags())
	return cmd
}

func newDecodeCmd() *cobra.Command {
	var out outputFlags
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
  leet decode --animate '|*455\/\/0|2) 4<<3|*73)'
  leet encode "Hello" | leet decode`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := alphabet()
			if err != nil {
				return err
			}
			return convert(cmd, args, out, a.Decode)
		},
	}
	out.register(cmd.Flags())
	return cmd
}

func convert(cmd *cobra.Command, args []string, o outputFlags, fn func(string) string) error {
	var copied []string
	emit := func(line string) error {
		if o.copy {
			copied = append(copied, line)
		}
		if o.animate {
			if w, ok := animatable(cmd.OutOrStdout(), line); ok {
				return animateLine(w, line, animFrames, animDelay)
			}
		}
		_, err := fmt.Fprintln(cmd.OutOrStdout(), line)
		return err
	}

	if len(args) > 0 {
		if err := emit(fn(strings.Join(args, " "))); err != nil {
			return err
		}
	} else {
		if isTerminal(cmd.InOrStdin()) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Type text and press Enter (Ctrl-D to finish):")
		}
		scanner := bufio.NewScanner(cmd.InOrStdin())
		for scanner.Scan() {
			if err := emit(fn(scanner.Text())); err != nil {
				return err
			}
		}
		if err := scanner.Err(); err != nil {
			return err
		}
	}
	if o.copy {
		return copyToClipboard(strings.Join(copied, "\n"))
	}
	return nil
}

const animFPS = 30

var (
	animFrames = int(anim.Duration.Seconds() * animFPS)
	animDelay  = time.Second / animFPS
)

// animatable reports whether line can be redrawn in place on w: w must be a
// terminal and the line must fit its width, since a wrapped line breaks \r.
func animatable(w io.Writer, line string) (io.Writer, bool) {
	f, ok := w.(*os.File)
	if !ok || !term.IsTerminal(int(f.Fd())) {
		return nil, false
	}
	width, _, err := term.GetSize(int(f.Fd()))
	return f, err == nil && utf8.RuneCountInString(line) < width
}

// animateLine redraws line in place, scrambled glyphs in green, then prints
// it for good.
func animateLine(w io.Writer, line string, frames int, delay time.Duration) error {
	seed := uint64(time.Now().UnixNano())
	var b strings.Builder
	for f := range frames {
		b.Reset()
		b.WriteByte('\r')
		for _, c := range anim.Cells(line, float64(f)/float64(frames), seed) {
			if c.Locked {
				b.WriteRune(c.R)
			} else {
				b.WriteString("\x1b[32m")
				b.WriteRune(c.R)
				b.WriteString("\x1b[0m")
			}
		}
		b.WriteString("\x1b[K")
		if _, err := io.WriteString(w, b.String()); err != nil {
			return err
		}
		time.Sleep(delay)
	}
	_, err := fmt.Fprintf(w, "\r%s\x1b[K\n", line)
	return err
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
