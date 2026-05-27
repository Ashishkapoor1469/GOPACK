package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"
	"github.com/Ashishkapoor1469/GOPACK/pkg/script"

	"github.com/spf13/cobra"
)

var (
	listScripts bool
	parallel    bool
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run [script]",
	Short: "Run package manifest scripts",
	Long:  `Run user-defined scripts declared in the gopack.json scripts block.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		mgr := installer.NewInstallManager(cfg)
		err = mgr.LoadManifest()
		if err != nil {
			fmt.Printf("Error loading manifest: %v\n", err)
			os.Exit(1)
		}

		scriptsMap := mgr.ManifestScripts()
		if listScripts {
			if len(scriptsMap) == 0 {
				fmt.Println("No scripts found in manifest.")
				return
			}
			fmt.Println("Available scripts:")
			for k, v := range scriptsMap {
				fmt.Printf("  %-15s : %s\n", k, v)
			}
			return
		}

		if len(args) == 0 {
			fmt.Println("Error: script name required. Run 'gpack run --list' to see available scripts.")
			os.Exit(1)
		}

		runner := script.NewScriptRunner(scriptsMap)

		if parallel {
			err = runner.RunParallel(args)
		} else {
			// Run single or sequential scripts
			for _, name := range args {
				err = runner.Run(name)
				if err != nil {
					break
				}
			}
		}

		if err != nil {
			fmt.Printf("\033[31mScript execution failed: %v\033[0m\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	runCmd.Flags().BoolVar(&listScripts, "list", false, "List all available scripts in gopack.json")
	runCmd.Flags().BoolVar(&parallel, "parallel", false, "Run multiple scripts in parallel")
}
