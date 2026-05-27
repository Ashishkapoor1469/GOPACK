package cmd

import (
	"fmt"
	"os"

	"github.com/Ashishkapoor1469/GOPACK/pkg/config"
	"github.com/Ashishkapoor1469/GOPACK/pkg/graph"
	"github.com/Ashishkapoor1469/GOPACK/pkg/installer"

	"github.com/spf13/cobra"
)

var (
	dotExport bool
	maxDepth  int
)

// graphCmd represents the graph command
var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "Render dependency tree",
	Long:  `Render a complete ASCII/Unicode dependency tree of installed packages or export to DOT format.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		mgr := installer.NewInstallManager(cfg)
		mgr.LoadManifest()

		// Read direct dependencies
		directDeps := mgr.ManifestDependencies()
		if len(directDeps) == 0 {
			// Try reading from lockfile instead if manifest dependencies is empty
			mgr.LoadLockfile()
			for k, v := range mgr.Dependencies() {
				directDeps[k] = v.Version
			}
		}

		if len(directDeps) == 0 {
			fmt.Println("No dependencies found. Install packages first.")
			return
		}

		tree := graph.BuildTree(directDeps)

		if dotExport {
			fmt.Println(graph.RenderDOT(tree))
			return
		}

		fmt.Println("Dependency Graph:")
		fmt.Print(graph.RenderTree(tree, 0, []string{}, nil))
	},
}

func init() {
	RootCmd.AddCommand(graphCmd)
	graphCmd.Flags().BoolVar(&dotExport, "dot", false, "Export to Graphviz DOT format")
	graphCmd.Flags().IntVar(&maxDepth, "depth", 5, "Max depth of the dependency tree to display")
}
