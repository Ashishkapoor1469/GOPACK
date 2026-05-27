package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/audit"
	"github.com/Ashishkapoor1469/GOPACK/pkg/tui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	offline bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "gp",
	Short: "GoPack is a high-performance CLI package manager.",
	Long:  `A production-grade package manager for JS/TS packages and full-stack frameworks written in Go.`,
	Run: func(cmd *cobra.Command, args []string) {
		// If there are arguments and they were not captured by subcommands,
		// launch TUI or print usage. (Normally rewritten by Execute() to direct install).
		if len(args) > 0 {
			// If we got here, it's either gp / or unknown.
			if args[0] == "/" {
				runTUI(true)
				return
			}
		}
		runInteractiveCLI()
	},
}

func runTUI(focusSearch bool) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	auditor, err := audit.NewAuditor(cfg)
	if err != nil {
		fmt.Printf("Error initializing auditor: %v\n", err)
		os.Exit(1)
	}
	defer auditor.Close()

	m := tui.NewTUIModel(cfg, auditor)
	if focusSearch {
		// Focus search immediately
		m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	// Intercept and preprocess arguments before Cobra parses them:
	// 1. If args contains no subcommands, run TUI (handled by Cobra RootCmd Run).
	// 2. If first argument is "/" -> keep it so RootCmd Run can open TUI with search focused.
	// 3. If first argument is not a known subcommand and doesn't start with "-" (flag), rewrite it as "install".
	if len(os.Args) > 1 {
		firstArg := os.Args[1]
		if firstArg != "/" && firstArg != "help" && !isKnownSubcommand(firstArg) && firstArg[0] != '-' {
			// Rewrite os.Args: inject "install" at position 1
			newArgs := make([]string, len(os.Args)+1)
			newArgs[0] = os.Args[0]
			newArgs[1] = "install"
			copy(newArgs[2:], os.Args[1:])
			os.Args = newArgs
		}
	}

	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func isKnownSubcommand(arg string) bool {
	known := map[string]bool{
		"install":       true, "i": true,
		"remove":        true, "rm": true,
		"search":        true, "s": true,
		"update":        true, "up": true,
		"list":          true,
		"audit":         true,
		"create":        true,
		"doctor":        true,
		"graph":         true,
		"health":        true,
		"run":           true,
		"licenses":      true,
		"config":        true,
		"cache":         true,
		"pin":           true,
		"unpin":         true,
		"why":           true,
		"outdated":      true,
		"verify":        true,
		"ws":            true,
		"global":        true,
		"shell":         true,
		"watch":         true,
		"patch":         true,
		"export":        true,
		"migrate":       true,
		"ci":            true,
		"env":           true,
		"prefetch":      true,
		"export-bundle": true,
		"import-bundle": true,
	}
	return known[arg]
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	}
	_, _ = config.LoadConfig()
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gopack/config.toml)")
	RootCmd.PersistentFlags().BoolVar(&offline, "offline", false, "force offline mode resolving from cache store")
}
