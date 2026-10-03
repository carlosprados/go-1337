package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

func newTableCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "table",
		Aliases: []string{"map", "alphabet"},
		Short:   "Show every letter with its 1337 variants and level",
		Long: `Print the alphabet: each letter, the lowest level that converts it,
its primary variant (used by default) and the alternatives --random may pick.
Reflects the custom alphabet when --map or the default config file is in use.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			a, err := alphabet()
			if err != nil {
				return err
			}
			t := table.New().
				Border(lipgloss.RoundedBorder()).
				BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("8"))).
				Headers("LETTER", "LEVEL", "PRIMARY", "ALTERNATIVES").
				StyleFunc(func(row, col int) lipgloss.Style {
					s := lipgloss.NewStyle().Padding(0, 1)
					switch {
					case row == table.HeaderRow:
						return s.Bold(true).Foreground(lipgloss.Color("10"))
					case col == 2:
						return s.Foreground(lipgloss.Color("10"))
					}
					return s
				})
			for _, e := range a.Table() {
				t.Row(string(e.Letter), e.Level.String(), e.Variants[0], strings.Join(e.Variants[1:], "  "))
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), t.Render())
			return err
		},
	}
}
