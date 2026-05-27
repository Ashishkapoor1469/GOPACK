package cmd

import (
	"fmt"
	"os"

	"gopack/pkg/config"
	"gopack/pkg/installer"
	"gopack/pkg/licenses"

	"github.com/spf13/cobra"
)

var checkPolicy string

// licensesCmd represents the licenses command
var licensesCmd = &cobra.Command{
	Use:   "licenses",
	Short: "Check licenses of dependencies",
	Long:  `Scan node_modules and output reports on license usage and compliance with the licensePolicy.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		mgr := installer.NewInstallManager(cfg)
		mgr.LoadManifest()

		allowed, denied := mgr.LicensePolicy()
		results, err := licenses.CheckCompliance(allowed, denied)
		if err != nil {
			fmt.Printf("\033[31mError scanning licenses: %v\033[0m\n", err)
			os.Exit(1)
		}

		if len(results) == 0 {
			fmt.Println("No package licenses found. Make sure packages are installed first.")
			return
		}

		fmt.Println("SPDX License Compliance Report:")
		fmt.Printf("%-25s %-15s %s\n", "PACKAGE", "LICENSE", "STATUS")
		fmt.Println(string(make([]byte, 55))) // separator

		hasViolation := false
		for _, r := range results {
			statusStr := "\033[32mALLOWED\033[0m"
			if r.Status == "DENIED" {
				statusStr = "\033[31mDENIED\033[0m"
				hasViolation = true
			} else if r.Status == "WARNING" {
				statusStr = "\033[33mWARNING\033[0m"
			}
			fmt.Printf("%-25s %-15s %s\n", r.PackageName, r.LicenseType, statusStr)
		}

		if checkPolicy == "strict" && hasViolation {
			fmt.Println("\n\033[31mError: License compliance check failed (found denied licenses).\033[0m")
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(licensesCmd)
	licensesCmd.Flags().StringVar(&checkPolicy, "check", "", "Policy enforcement check (e.g. strict)")
}
