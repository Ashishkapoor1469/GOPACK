package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"
	"github.com/Ashishkapoor1469/GOPACK/pkg/registry"
	"github.com/Ashishkapoor1469/GOPACK/pkg/audit"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type tab int

const (
	tabFramework tab = iota
	tabPackages
)

// Msg types
type searchFinishedMsg struct {
	results []registry.SearchPackageInfo
	err     error
}

type installProgressMsg struct {
	progress []installer.ProgressInfo
}

type installFinishedMsg struct {
	err error
}

type TUIModel struct {
	cfg            *config.Config
	client         *registry.RegistryClient
	auditor        *audit.Auditor
	activeTab      tab
	inputs         [2]textinput.Model
	frameworks     []string
	filteredFws    []string
	pkgResults     []registry.SearchPackageInfo
	cursorIdx      int
	selectedPkgs   map[string]bool // Holds Space-selected packages
	installing     bool
	installError   string
	progressList   []installer.ProgressInfo
	quitting       bool
	terminalWidth  int
	terminalHeight int
	infoText       string
}

func NewTUIModel(cfg *config.Config, auditor *audit.Auditor) TUIModel {
	fwInput := textinput.New()
	fwInput.Placeholder = "search frameworks..."
	fwInput.Focus()
	fwInput.Prompt = "> "
	fwInput.CharLimit = 50

	pkgInput := textinput.New()
	pkgInput.Placeholder = "search packages..."
	pkgInput.Prompt = "> "
	pkgInput.CharLimit = 50

	fws := []string{
		"Next.js", "NestJS", "React", "Vue 3", "Svelte", "SvelteKit", "Astro",
		"Remix", "Nuxt 3", "Angular", "Solid.js", "Qwik", "Expo", "Electron", "Tauri",
	}

	return TUIModel{
		cfg:          cfg,
		client:       registry.NewRegistryClient(),
		auditor:      auditor,
		activeTab:    tabFramework,
		inputs:       [2]textinput.Model{fwInput, pkgInput},
		frameworks:   fws,
		filteredFws:  fws,
		selectedPkgs: make(map[string]bool),
	}
}

func (m TUIModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil

	case searchFinishedMsg:
		if msg.err == nil {
			m.pkgResults = msg.results
		}
		return m, nil

	case installProgressMsg:
		m.progressList = msg.progress
		return m, nil

	case installFinishedMsg:
		m.installing = false
		if msg.err != nil {
			m.installError = msg.err.Error()
		} else {
			m.selectedPkgs = make(map[string]bool)
			m.infoText = "Installation completed successfully!"
			go func() {
				time.Sleep(3 * time.Second)
				m.infoText = ""
			}()
		}
		return m, nil

	case tea.KeyMsg:
		if m.installing {
			return m, nil // Block input during installation
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "tab", "right", "left":
			// Toggle between Frameworks and Packages tabs
			m.inputs[m.activeTab].Blur()
			if msg.String() == "left" || msg.String() == "shift+tab" {
				m.activeTab = tabFramework
			} else {
				m.activeTab = tabPackages
			}
			m.inputs[m.activeTab].Focus()
			m.cursorIdx = 0
			return m, nil

		case "up":
			if m.cursorIdx > 0 {
				m.cursorIdx--
			}
			return m, nil

		case "down":
			max := 0
			if m.activeTab == tabFramework {
				max = len(m.filteredFws)
			} else {
				max = len(m.pkgResults)
			}
			if m.cursorIdx < max-1 {
				m.cursorIdx++
			}
			return m, nil

		case " ":
			// Toggle selection
			var name string
			if m.activeTab == tabFramework {
				if m.cursorIdx < len(m.filteredFws) {
					name = m.filteredFws[m.cursorIdx]
				}
			} else {
				if m.cursorIdx < len(m.pkgResults) {
					name = m.pkgResults[m.cursorIdx].Name
				}
			}

			if name != "" {
				if m.selectedPkgs[name] {
					delete(m.selectedPkgs, name)
				} else {
					m.selectedPkgs[name] = true
				}
			}
			return m, nil

		case "enter":
			// Run installation
			var listToInstall []string
			if len(m.selectedPkgs) > 0 {
				for p := range m.selectedPkgs {
					listToInstall = append(listToInstall, p)
				}
			} else {
				// Install highlighted item
				if m.activeTab == tabFramework {
					if m.cursorIdx < len(m.filteredFws) {
						listToInstall = append(listToInstall, m.filteredFws[m.cursorIdx])
					}
				} else {
					if m.cursorIdx < len(m.pkgResults) {
						listToInstall = append(listToInstall, m.pkgResults[m.cursorIdx].Name)
					}
				}
			}

			if len(listToInstall) > 0 {
				m.installing = true
				m.installError = ""
				return m, m.triggerInstall(listToInstall)
			}
			return m, nil

		case "/":
			// Refocus / clear search
			m.inputs[m.activeTab].SetValue("")
			m.inputs[m.activeTab].Focus()
			m.cursorIdx = 0
			m.filterList()
			return m, nil

		case "esc":
			// Clear selection queue
			m.selectedPkgs = make(map[string]bool)
			return m, nil
		}
	}

	// Handle input typing
	prevVal := m.inputs[m.activeTab].Value()
	var cmd tea.Cmd
	m.inputs[m.activeTab], cmd = m.inputs[m.activeTab].Update(msg)
	
	if m.inputs[m.activeTab].Value() != prevVal {
		m.cursorIdx = 0
		m.filterList()
		if m.activeTab == tabPackages {
			// Trigger debounced registry search
			return m, m.searchRegistry(m.inputs[tabPackages].Value())
		}
	}

	return m, cmd
}

func (m *TUIModel) filterList() {
	val := m.inputs[tabFramework].Value()
	if val == "" {
		m.filteredFws = m.frameworks
		return
	}

	var filtered []string
	for _, f := range m.frameworks {
		if strings.Contains(strings.ToLower(f), strings.ToLower(val)) {
			filtered = append(filtered, f)
		}
	}
	m.filteredFws = filtered
}

func (m TUIModel) searchRegistry(query string) tea.Cmd {
	return tea.Tick(300*time.Millisecond, func(time.Time) tea.Msg {
		results, err := m.client.SearchPackages(query, 20)
		return searchFinishedMsg{results: results, err: err}
	})
}

func (m TUIModel) triggerInstall(pkgs []string) tea.Cmd {
	return func() tea.Msg {
		mgr := installer.NewInstallManager(m.cfg)
		
		// Build maps
		targets := make(map[string]string)
		for _, p := range pkgs {
			targets[p] = "latest"
		}
		
		resolved, err := mgr.ResolveDependencyTree(targets)
		if err != nil {
			return installFinishedMsg{err: err}
		}

		progressChan := make(chan []installer.ProgressInfo)
		done := make(chan error)

		go func() {
			done <- mgr.InstallResolvedPackages(resolved, progressChan, false)
		}()

		// Consume progress channel in separate loop to notify Bubble Tea
		go func() {
			for range progressChan {
				// Discard/drain progress channel
			}
		}()

		err = <-done
		return installFinishedMsg{err: err}
	}
}

func (m TUIModel) View() string {
	if m.quitting {
		return "Thanks for using GoPack!\n"
	}

	// Themes
	purple := lipgloss.Color("#7F77DD")
	gray := lipgloss.Color("#6272A4")
	lightGray := lipgloss.Color("#A0A0A0")
	green := lipgloss.Color("#50FA7B")
	yellow := lipgloss.Color("#F1FA8C")
	red := lipgloss.Color("#FF5555")

	// Custom styles
	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Padding(1, 2)
	activeTabStyle := lipgloss.NewStyle().Border(lipgloss.NormalBorder(), false, false, true, false).BorderForeground(purple).Foreground(purple).Bold(true).Padding(0, 1)
	inactiveTabStyle := lipgloss.NewStyle().Foreground(gray).Padding(0, 1)
	titleStyle := lipgloss.NewStyle().Foreground(gray).Italic(true)
	
	// Direct mode rendering
	if m.installing {
		var sb strings.Builder
		sb.WriteString("┌─ installing packages ──────────────────────────────┐\n")
		if len(m.progressList) == 0 {
			sb.WriteString("│  Resolving and downloading...                      │\n")
		}
		for _, p := range m.progressList {
			statusSymbol := "░"
			if p.Status == "done" {
				statusSymbol = "✔"
			} else if p.Status == "failed" {
				statusSymbol = "✘"
			} else {
				statusSymbol = "●"
			}
			barLength := 20
			filledLength := int(p.Percent / 100 * float64(barLength))
			bar := strings.Repeat("█", filledLength) + strings.Repeat("░", barLength-filledLength)
			sb.WriteString(fmt.Sprintf("│  %s %-18s  %s  %3.0f%%  │\n", statusSymbol, p.PackageName, bar, p.Percent))
		}
		sb.WriteString("└────────────────────────────────────────────────────┘\n")
		return sb.String()
	}

	var view strings.Builder

	// Top Bar
	view.WriteString(titleStyle.Render(fmt.Sprintf("gopack v2.0.0  ·  registry: %s  ·  store: %s  ·  offline-ready\n\n", "registry.npmjs.org", "~/.gopack/store")))

	// Side-by-side search panels
	fwTitle := inactiveTabStyle.Render("FRAMEWORK")
	pkgTitle := inactiveTabStyle.Render("PACKAGES")
	if m.activeTab == tabFramework {
		fwTitle = activeTabStyle.Render("FRAMEWORK")
	} else {
		pkgTitle = activeTabStyle.Render("PACKAGES")
	}

	// Render Search Boxes
	fwBoxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gray).Width(30).Padding(0, 1)
	pkgBoxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gray).Width(30).Padding(0, 1)
	
	if m.activeTab == tabFramework {
		fwBoxStyle = fwBoxStyle.BorderForeground(purple)
	} else {
		pkgBoxStyle = pkgBoxStyle.BorderForeground(purple)
	}

	searchPanels := lipgloss.JoinHorizontal(
		lipgloss.Top,
		lipgloss.JoinVertical(lipgloss.Left, fwTitle, fwBoxStyle.Render(m.inputs[tabFramework].View())),
		"   │   ",
		lipgloss.JoinVertical(lipgloss.Left, pkgTitle, pkgBoxStyle.Render(m.inputs[tabPackages].View())),
	)
	view.WriteString(searchPanels + "\n\n")

	// Results Header
	view.WriteString(lipgloss.NewStyle().Foreground(gray).Render("RESULTS · Space=select · Enter=install\n"))

	// Results List
	var resultsSb strings.Builder
	if m.activeTab == tabFramework {
		for i, fw := range m.filteredFws {
			cursor := "  "
			if i == m.cursorIdx {
				cursor = "● "
			}

			selected := "○"
			if m.selectedPkgs[fw] {
				selected = "☑"
			}

			// Render Row
			rowStyle := lipgloss.NewStyle().Foreground(lightGray)
			if i == m.cursorIdx {
				rowStyle = rowStyle.Foreground(purple).Background(lipgloss.Color("#262335"))
			}

			badge := lipgloss.NewStyle().Background(lipgloss.Color("#1b4d22")).Foreground(green).Render(" ● safe ")

			row := fmt.Sprintf("%s %s %-20s  %-30s  %s", cursor, selected, fw, "Boilerplate scaffolding", badge)
			resultsSb.WriteString(rowStyle.Render(row) + "\n")
		}
	} else {
		for i, pkg := range m.pkgResults {
			cursor := "  "
			if i == m.cursorIdx {
				cursor = "● "
			}

			selected := "○"
			if m.selectedPkgs[pkg.Name] {
				selected = "☑"
			}

			rowStyle := lipgloss.NewStyle().Foreground(lightGray)
			if i == m.cursorIdx {
				rowStyle = rowStyle.Foreground(purple).Background(lipgloss.Color("#262335"))
			}

			// Mock badge
			badge := lipgloss.NewStyle().Background(lipgloss.Color("#1b4d22")).Foreground(green).Render(" ● safe ")
			if strings.Contains(pkg.Name, "lodash") {
				badge = lipgloss.NewStyle().Background(lipgloss.Color("#4d1b1b")).Foreground(red).Render(" ● CVE ")
			} else if strings.Contains(pkg.Name, "express") {
				badge = lipgloss.NewStyle().Background(lipgloss.Color("#4d3f1b")).Foreground(yellow).Render(" ● low ")
			}

			// Format weekly downloads
			dls := fmt.Sprintf("%dM/wk", pkg.WeeklyDownloads/1000000)
			if pkg.WeeklyDownloads < 1000000 {
				dls = fmt.Sprintf("%dK/wk", pkg.WeeklyDownloads/1000)
			}

			desc := pkg.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}

			row := fmt.Sprintf("%s %s %-20s  %-10s  %-35s  %-8s %s", cursor, selected, pkg.Name, pkg.Version, desc, dls, badge)
			resultsSb.WriteString(rowStyle.Render(row) + "\n")
		}
	}

	view.WriteString(resultsSb.String() + "\n")

	// Selected Queue Display
	if len(m.selectedPkgs) > 0 {
		var queueSb strings.Builder
		queueSb.WriteString(lipgloss.NewStyle().Foreground(green).Render("INSTALL QUEUE · Enter to install all\n"))
		for p := range m.selectedPkgs {
			queueSb.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("#233526")).Foreground(green).Padding(0, 1).Render(p+" ✕") + "  ")
		}
		view.WriteString(queueSb.String() + "\n\n")
	}

	if m.infoText != "" {
		view.WriteString(lipgloss.NewStyle().Foreground(green).Render(m.infoText) + "\n\n")
	}
	if m.installError != "" {
		view.WriteString(lipgloss.NewStyle().Foreground(red).Render("Error: "+m.installError) + "\n\n")
	}

	// Status Bar
	statusBar := lipgloss.NewStyle().Foreground(gray).Render("[↑↓] navigate  ·  [←→] switch tab  ·  [Space] select  ·  [Enter] install  ·  [/] clear  ·  [q] quit")
	view.WriteString(statusBar)

	return borderStyle.Render(view.String())
}
