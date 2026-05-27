package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/audit"
	"github.com/Ashishkapoor1469/GOPACK/pkg/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:     "search [query]",
	Aliases: []string{"s"},
	Short:   "Search packages",
	Long:    `Search npm packages using interactive terminal interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		auditor, err := audit.NewAuditor(cfg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		defer auditor.Close()

		m := tui.NewTUIModel(cfg, auditor)
		if len(args) > 0 {
			// Focus packages tab
			m.Update(tea.KeyMsg{Type: tea.KeyTab}) // Go to packages tab
			m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}}) // Focus search
			// Feed query characters
			for _, char := range args[0] {
				m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{char}})
			}
		}

		p := tea.NewProgram(m, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Error running TUI: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(searchCmd)
}
