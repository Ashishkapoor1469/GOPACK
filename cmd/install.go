package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"

	"github.com/spf13/cobra"
)

var (
	envProfile string
	saveDev    bool
	frozen     bool
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:     "install [packages]",
	Aliases: []string{"i"},
	Short:   "Install package dependencies",
	Long:    `Install latest or specific versions of packages and update gopack.json/gopack.lock.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		mgr := installer.NewInstallManager(cfg)

		isDev := saveDev || envProfile == "dev"
		
		fmt.Printf("GoPack: Resolving and installing packages...\n")
		err = mgr.Install(args, isDev)
		if err != nil {
			fmt.Printf("\r\n\033[31mError during install: %v\033[0m\n", err)
			os.Exit(1)
		}
		fmt.Printf("\r\n\033[32mSuccess: Packages installed successfully!\033[0m\n")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
	installCmd.Flags().StringVar(&envProfile, "env", "production", "Install environment profile (dev/production)")
	installCmd.Flags().BoolVarP(&saveDev, "save-dev", "D", false, "Save package to devDependencies")
	installCmd.Flags().BoolVar(&frozen, "frozen", false, "Refuse to change lockfile; fail if it would change")
}
