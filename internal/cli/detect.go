package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDetectCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "detect [text...]",
		Aliases: []string{"score"},
		Short:   "Score how much of a text is written in 1337",
		Long: `Score how much of a text is written in 1337 and show its decoded form.

The score is the share of letters written as leet sequences. Digits and
symbols in ordinary prose count as leet too, so take numbers-heavy text
with a grain of salt.

With arguments, they are joined with spaces. Without arguments, the whole of
standard input is scored as one text.`,
		Example: `  leet detect 'h4ck th3 p14n37'
  leet detect < suspicious.txt`,
		RunE: func(cmd *cobra.Command, args []string) error {
			a, err := alphabet()
			if err != nil {
				return err
			}
			text, err := readInput(cmd, args)
			if err != nil {
				return err
			}
			s := a.Detect(text)
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "score    %.0f%% (%s)\ndecoded  %s\n", s.Ratio*100, s.Verdict(), s.Decoded)
			return err
		},
	}
}
