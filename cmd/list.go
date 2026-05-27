package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/audit"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed package dependencies",
	Long:  `Show all direct and transitive packages currently installed in the project with their security status.`,
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

		mgr := installer.NewInstallManager(cfg)
		mgr.LoadLockfile()

		deps := mgr.Dependencies()
		if len(deps) == 0 {
			fmt.Println("No package dependencies found in lockfile. Run 'gp install' first.")
			return
		}

		fmt.Println("Installed Dependencies:")
		fmt.Printf("%-25s %-15s %s\n", "NAME", "VERSION", "SECURITY")
		fmt.Println("-------------------------------------------------------------") // horizontal bar separator
		for name, dep := range deps {
			res, err := auditor.AuditPackage(name, dep.Version, offline)
			badge := "\033[32m● safe\033[0m"
			if err == nil {
				if res.Severity == audit.Red {
					badge = "\033[31m● CVE (critical)\033[0m"
				} else if res.Severity == audit.Yellow {
					badge = "\033[33m● low advisory\033[0m"
				}
			}
			fmt.Printf("%-25s %-15s %s\n", name, dep.Version, badge)
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
