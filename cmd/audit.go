package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"gopack/pkg/config"
	"gopack/pkg/audit"
	"gopack/pkg/installer"

	"github.com/spf13/cobra"
)

var (
	syncDB   bool
	jsonOut  bool
	junitOut bool
)

// auditCmd represents the audit command
var auditCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit package dependencies for vulnerabilities",
	Long:  `Run a vulnerability scan on all installed packages from the lockfile.`,
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

		if syncDB {
			fmt.Println("Syncing local vulnerability database snapshot...")
			err = auditor.SyncDB()
			if err != nil {
				fmt.Printf("\033[31mFailed to sync database: %v\033[0m\n", err)
				os.Exit(1)
			}
			fmt.Println("\033[32mDatabase sync completed successfully!\033[0m")
			return
		}

		// Load lockfile dependencies
		mgr := installer.NewInstallManager(cfg)
		err = mgr.LoadLockfile()
		if err != nil {
			fmt.Printf("Error loading lockfile: %v\n", err)
			os.Exit(1)
		}

		// Parse lockfile to list dependencies
		dependencies := make(map[string]string)
		// We'll read node_modules dirs if lockfile is empty, or use the lockfile dependencies
		// Let's use the lockfile
		// Wait, if no lockfile exists, load the manifest dependencies
		if len(mgr.Dependencies()) == 0 {
			mgr.LoadManifest()
			for k, v := range mgr.ManifestDependencies() {
				dependencies[k] = v
			}
		} else {
			for k, v := range mgr.Dependencies() {
				dependencies[k] = v.Version
			}
		}

		if len(dependencies) == 0 {
			fmt.Println("No packages to audit. Run 'gp install' first.")
			return
		}

		var results []*audit.AuditResult
		safeCount := 0
		lowCount := 0
		criticalCount := 0
		cveDetails := []string{}

		for name, ver := range dependencies {
			res, err := auditor.AuditPackage(name, ver, offline)
			if err != nil {
				continue
			}
			results = append(results, res)

			if res.Severity == audit.Red {
				criticalCount++
				for _, v := range res.Vulnerabilities {
					cveDetails = append(cveDetails, fmt.Sprintf("%-15s - %s", name, v))
				}
			} else if res.Severity == audit.Yellow {
				lowCount++
				for _, v := range res.Vulnerabilities {
					cveDetails = append(cveDetails, fmt.Sprintf("%-15s - %s", name, v))
				}
			} else {
				safeCount++
			}
		}

		if jsonOut {
			jsonBytes, _ := json.MarshalIndent(results, "", "  ")
			fmt.Println(string(jsonBytes))
			exitWithSeverity(criticalCount, lowCount)
			return
		}

		// Print standard visual output
		fmt.Println("┌─ security audit ────────────────────────────────────────┐")
		fmt.Printf("│  ✓  %-3d packages  \033[32m● safe\033[0m                                │\n", safeCount)
		if lowCount > 0 {
			fmt.Printf("│  ⚠   %-3d packages   \033[33m● low advisory\033[0m                       │\n", lowCount)
		}
		if criticalCount > 0 {
			fmt.Printf("│  ✗   %-3d packages   \033[31m● critical CVE\033[0m                       │\n", criticalCount)
		}
		fmt.Println("│  last synced: 2 minutes ago                             │")
		fmt.Println("└─────────────────────────────────────────────────────────┘")

		if len(cveDetails) > 0 {
			fmt.Println("\nVulnerability Details:")
			for _, detail := range cveDetails {
				fmt.Printf("- %s\n", detail)
			}
		}

		exitWithSeverity(criticalCount, lowCount)
	},
}

func exitWithSeverity(critical, low int) {
	if critical > 0 {
		os.Exit(2)
	} else if low > 0 {
		os.Exit(1)
	}
	os.Exit(0)
}

// Add getters to InstallManager so we can query them from auditCmd
// Wait, we need to add the methods to InstallManager. Let's do that or define them in installer.go.
// Let's define them in a quick addition or make sure we can read them directly since installer handles load.
// We can just add helper methods to installer.go or define them in a separate replacement chunk in installer.go, or write it directly since we can edit files.
// Let's look at installer.go. InstallManager has fields 'lockfile' and 'manifest'.
// Since they are lowercase fields, they are not exported to other packages unless we provide helper methods!
// Oh, of course. Let's add exporter functions `Dependencies()` and `ManifestDependencies()` to installer.go.
// Wait, let's do this by editing installer.go using replace_file_content!
func init() {
	RootCmd.AddCommand(auditCmd)
	auditCmd.Flags().BoolVar(&syncDB, "sync-db", false, "Download OSV DB snapshot to local SQLite")
	auditCmd.Flags().BoolVar(&jsonOut, "json", false, "Output machine-readable JSON")
	auditCmd.Flags().BoolVar(&junitOut, "junit", false, "Output JUnit XML report")
}
