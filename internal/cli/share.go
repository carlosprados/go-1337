package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/carlosprados/go-1337/internal/leet"
	"github.com/spf13/cobra"
)

const shareBaseURL = "https://carlosprados.github.io/go-1337/"

// shareURL links to the web demo with leetText preloaded in decode mode, so
// whoever opens it watches the message decrypt.
func shareURL(leetText string) string {
	return shareBaseURL + "#" + url.Values{"d": {leetText}}.Encode()
}

func newShareCmd() *cobra.Command {
	var (
		ef     encodeFlags
		toClip bool
	)
	cmd := &cobra.Command{
		Use:   "share [text...]",
		Short: "Print a web link that decrypts your message on screen",
		Long: `Encode a message and print a link to the web demo that carries it in 1337.
Whoever opens the link watches the message decrypt in their browser.

The link holds only the leet text, not the original. The web demo decodes with
the built-in alphabet, so a custom --map is ignored here.

With arguments, they are joined with spaces. Without arguments, the whole of
standard input is shared as one message.`,
		Example: `  leet share "meet me at the usual place"
  leet share --random -c "the cake is a lie"
  fortune | leet share`,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts, err := ef.options()
			if err != nil {
				return err
			}
			text, err := readInput(cmd, args)
			if err != nil {
				return err
			}
			link := shareURL(leet.Default().Encode(strings.TrimRight(text, "\n"), opts))
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), link); err != nil {
				return err
			}
			if toClip {
				return copyToClipboard(link)
			}
			return nil
		},
	}
	ef.register(cmd)
	cmd.Flags().BoolVarP(&toClip, "copy", "c", false, "also copy the link to the clipboard")
	return cmd
}
