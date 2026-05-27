package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"

	"golang.org/x/term"
)

var menuOptions = []struct {
	cmd  string
	desc string
}{
	{"/search",   "Open interactive TUI search"},
	{"/health",   "View project health dashboard"},
	{"/audit",    "Scan dependencies for vulnerabilities"},
	{"/graph",    "Render dependency tree"},
	{"/licenses", "Check license compliance"},
	{"/help",     "Show command help"},
	{"/exit",     "Exit GoPack CLI"},
}

// runInteractiveCLI launches the Claude Code-like CLI loop
func runInteractiveCLI() {
	// Print Banner
	printBanner()

	// Initialize terminal in raw mode
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Error entering raw terminal: %v\n", err)
		return
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	buffer := ""
	menuOpen := false
	selectedMenuIdx := 0
	lastMenuLinesCount := 0

	// Helper to restore terminal temporarily to run subcommands and then re-enter raw mode
	runInShellMode := func(f func()) {
		term.Restore(int(os.Stdin.Fd()), oldState)
		fmt.Println() // print newline
		f()
		// Re-enter raw mode
		var err error
		oldState, err = term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			fmt.Printf("Error re-entering raw terminal: %v\n", err)
			os.Exit(1)
		}
	}

	clearMenuLines := func() {
		if lastMenuLinesCount > 0 {
			// Move cursor down, clear lines, and move cursor back up
			for i := 0; i < lastMenuLinesCount; i++ {
				fmt.Print("\n\033[K")
			}
			// Move cursor up to prompt line
			fmt.Printf("\033[%dA", lastMenuLinesCount)
			lastMenuLinesCount = 0
		}
	}

	render := func() {
		// 1. Clear previous menu lines
		clearMenuLines()

		// 2. Print prompt line
		// Move cursor to beginning of prompt line and clear it
		fmt.Print("\r\033[K")
		fmt.Printf("\033[38;5;141mgp > \033[0m%s", buffer)

		// 3. Print menu if open
		if menuOpen {
			filtered := filterMenu(buffer)
			if len(filtered) > 0 {
				if selectedMenuIdx >= len(filtered) {
					selectedMenuIdx = len(filtered) - 1
				}
				if selectedMenuIdx < 0 {
					selectedMenuIdx = 0
				}

				fmt.Print("\n\033[K\033[38;5;242mSuggestions:\033[0m")
				lastMenuLinesCount = 1

				for i, opt := range filtered {
					if i == selectedMenuIdx {
						fmt.Printf("\n\033[K  \033[38;5;141m❯ %-10s \033[38;5;246m%s\033[0m", opt.cmd, opt.desc)
					} else {
						fmt.Printf("\n\033[K    %-10s \033[38;5;242m%s\033[0m", opt.cmd, opt.desc)
					}
					lastMenuLinesCount++
				}

				// Move cursor back up to prompt line and position it at the end of the text
				fmt.Printf("\033[%dA", lastMenuLinesCount)
				fmt.Printf("\r\033[%dC", len("gp > ")+len(buffer))
			} else {
				menuOpen = false
			}
		}
	}

	// Initial render
	render()

	reader := bufio.NewReader(os.Stdin)
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			break
		}

		// Handle ANSI Escape Sequences (Arrows, etc.)
		if r == '\x1b' {
			// Peek next bytes to see if it's an arrow key
			nextBytes := make([]byte, 2)
			n, _ := os.Stdin.Read(nextBytes)
			if n == 2 && nextBytes[0] == '[' {
				switch nextBytes[1] {
				case 'A': // Arrow Up
					if menuOpen {
						filtered := filterMenu(buffer)
						if selectedMenuIdx > 0 {
							selectedMenuIdx--
						} else {
							selectedMenuIdx = len(filtered) - 1
						}
						render()
					}
					continue
				case 'B': // Arrow Down
					if menuOpen {
						filtered := filterMenu(buffer)
						if selectedMenuIdx < len(filtered)-1 {
							selectedMenuIdx++
						} else {
							selectedMenuIdx = 0
						}
						render()
					}
					continue
				}
			}
			continue
		}

		// Handle Ctrl+C or Ctrl+D
		if r == '\x03' || r == '\x04' {
			clearMenuLines()
			term.Restore(int(os.Stdin.Fd()), oldState)
			fmt.Println("\nGoodbye!")
			return
		}

		// Handle Enter key
		if r == '\r' || r == '\n' {
			clearMenuLines()

			if menuOpen {
				// Autocomplete the selected option
				filtered := filterMenu(buffer)
				if len(filtered) > 0 && selectedMenuIdx < len(filtered) {
					buffer = filtered[selectedMenuIdx].cmd
					menuOpen = false
					render()
				}
				continue
			}

			// Run command
			cmdText := strings.TrimSpace(buffer)
			buffer = "" // reset buffer

			if cmdText == "" {
				render()
				continue
			}

			// Handle commands
			if cmdText == "/exit" || cmdText == "/quit" {
				term.Restore(int(os.Stdin.Fd()), oldState)
				fmt.Println("\nGoodbye!")
				return
			}

			if strings.HasPrefix(cmdText, "/") {
				// Run TUI/Subcommands
				runInShellMode(func() {
					switch cmdText {
					case "/search":
						runTUI(false)
					case "/health":
						runHealthCmd()
					case "/audit":
						runAuditCmd()
					case "/graph":
						runGraphCmd()
					case "/licenses":
						runLicensesCmd()
					case "/help":
						printHelp()
					default:
						fmt.Printf("Unknown command: %s. Type /help to see available commands.\n", cmdText)
					}
				})
			} else {
				// Treat as package install
				runInShellMode(func() {
					pkgs := strings.Fields(cmdText)
					if len(pkgs) > 0 {
						if pkgs[0] == "install" || pkgs[0] == "i" {
							pkgs = pkgs[1:]
						}
					}
					if len(pkgs) > 0 {
						cfg, _ := config.LoadConfig()
						mgr := installer.NewInstallManager(cfg)
						fmt.Printf("GoPack: Installing %s...\n", strings.Join(pkgs, ", "))
						err := mgr.Install(pkgs, false)
						if err != nil {
							fmt.Printf("\033[31mError installing packages: %v\033[0m\n", err)
						} else {
							fmt.Println("\033[32mSuccess: Packages installed successfully!\033[0m")
						}
					} else {
						fmt.Println("No package name entered. Type a package name (e.g. lodash) to install it.")
					}
				})
			}

			render()
			continue
		}

		// Handle Backspace
		if r == '\x7f' || r == '\x08' {
			if len(buffer) > 0 {
				buffer = buffer[:len(buffer)-1]
				if !strings.HasPrefix(buffer, "/") {
					menuOpen = false
				}
				render()
			}
			continue
		}

		// Handle Tab
		if r == '\t' {
			if menuOpen {
				filtered := filterMenu(buffer)
				if len(filtered) > 0 {
					buffer = filtered[selectedMenuIdx].cmd
					menuOpen = false
					render()
				}
			} else if strings.HasPrefix(buffer, "/") {
				menuOpen = true
				selectedMenuIdx = 0
				render()
			}
			continue
		}

		// Default: append character to input buffer
		buffer += string(r)
		if strings.HasPrefix(buffer, "/") {
			menuOpen = true
		}
		render()
	}
}

func filterMenu(input string) []struct{ cmd, desc string } {
	if input == "" || input == "/" {
		return menuOptions
	}

	var filtered []struct{ cmd, desc string }
	for _, opt := range menuOptions {
		if strings.HasPrefix(opt.cmd, input) {
			filtered = append(filtered, opt)
		}
	}
	return filtered
}

func printBanner() {
	banner := `
\033[38;5;141m  ______   ______   .______      ___       ______  __  ___ 
 /  ____| /  __  \  |   _  \    /   \     /  ____||  |/  / 
|  |  __ |  |  |  | |  |_)  |  /  ^  \   |  |     |  '  /  
|  | |_ | |  |  |  | |   ___/  /  /_\  \  |  |     |    <   
|  |__| | |  ` + "`" + `--'  | |  |     /  _____  \ |  ` + "`" + `----.|  .  \  
 \______|  \______/  |__|    /__/     \__\ \______||__|\__\ \033[0m

  \033[38;5;141mGoPack CLI v2.0.0\033[0m · registry: \033[36mregistry.npmjs.org\033[0m · store: \033[36m~/.gopack/store\033[0m
  Type a package name to install it (e.g. \033[36mlodash\033[0m) or type \033[35m/\033[0m to see suggestions.
`
	// Unquote backtick strings if needed or output directly
	fmt.Println(banner)
}

func printHelp() {
	fmt.Println("Available shell commands:")
	for _, opt := range menuOptions {
		fmt.Printf("  %-10s - %s\n", opt.cmd, opt.desc)
	}
	fmt.Println("\nOr type any npm package name directly (e.g. 'lodash' or 'express') to install it.")
}

// Redefine run loops for subcommands to make them callable from shell
func runHealthCmd() {
	cfg, _ := config.LoadConfig()
	mgr := installer.NewInstallManager(cfg)
	mgr.LoadManifest()
	mgr.LoadLockfile()

	if len(mgr.Dependencies()) == 0 {
		fmt.Println("No packages installed. Please run 'gp install' to set up dependencies first.")
		return
	}

	healthCmd.Run(healthCmd, nil)
}

func runAuditCmd() {
	cfg, _ := config.LoadConfig()
	mgr := installer.NewInstallManager(cfg)
	mgr.LoadLockfile()

	auditCmd.Run(auditCmd, nil)
}

func runGraphCmd() {
	graphCmd.Run(graphCmd, nil)
}

func runLicensesCmd() {
	licensesCmd.Run(licensesCmd, nil)
}
