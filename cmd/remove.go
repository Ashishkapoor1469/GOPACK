package cmd

import (
	"fmt"
	"os"

	"gopack/pkg/config"
	"gopack/pkg/installer"

	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:     "remove [package]",
	Aliases: []string{"rm"},
	Short:   "Uninstall package dependency",
	Long:    `Remove package dependency from manifest, lockfile and node_modules.`,
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		mgr := installer.NewInstallManager(cfg)
		pkgName := args[0]

		fmt.Printf("Removing package %s...\n", pkgName)
		err = mgr.Remove(pkgName)
		if err != nil {
			fmt.Printf("\033[31mError during package removal: %v\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Printf("\033[32mSuccess: Package %s removed successfully!\033[0m\n", pkgName)
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
