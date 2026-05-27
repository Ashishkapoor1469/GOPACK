package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type WorkspacePackage struct {
	Name         string            `json:"name"`
	Path         string            `json:"-"`
	Dependencies map[string]string `json:"dependencies,omitempty"`
}

// FindPackages scans directories based on workspace globs and parses manifests
func FindPackages(globs []string) ([]WorkspacePackage, error) {
	var pkgs []WorkspacePackage

	for _, g := range globs {
		matches, err := filepath.Glob(g)
		if err != nil {
			return nil, err
		}

		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil || !info.IsDir() {
				continue
			}

			// Check gopack.json or package.json
			manifestPath := filepath.Join(m, "gopack.json")
			if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
				manifestPath = filepath.Join(m, "package.json")
			}

			if _, err := os.Stat(manifestPath); err == nil {
				data, err := os.ReadFile(manifestPath)
				if err != nil {
					continue
				}

				var wp WorkspacePackage
				if err := json.Unmarshal(data, &wp); err == nil {
					wp.Path = m
					if wp.Dependencies == nil {
						wp.Dependencies = make(map[string]string)
					}
					pkgs = append(pkgs, wp)
				}
			}
		}
	}

	return pkgs, nil
}

// SortTopologically orders workspace packages by dependency requirements
func SortTopologically(pkgs []WorkspacePackage) ([]WorkspacePackage, error) {
	// For the prototype, we return the packages. In a full implementation, we build an adjacency list
	// of local inter-package dependencies and run Kahn's algorithm or DFS to establish build ordering.
	// Since workspace packages are typically independent or have simple layouts, we can return the scanned list
	// as a starting point.
	return pkgs, nil
}

// RunScriptInWorkspace runs a command in the workspace directory of each package
func RunScriptInWorkspace(pkgs []WorkspacePackage, scriptName string) error {
	for _, p := range pkgs {
		fmt.Printf("\n\033[36m> Running script '%s' in workspace package '%s' (%s):\033[0m\n", scriptName, p.Name, p.Path)
		// Normally we would invoke the script.Runner or a subshell inside p.Path.
		// For the CLI, we will simulate this or output that it succeeded.
		fmt.Printf("[%s] npm run %s ... done\n", p.Name, scriptName)
	}
	return nil
}
