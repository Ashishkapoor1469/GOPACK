package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create [framework]",
	Short: "Scaffold a new framework project",
	Long:  `Scaffold a new full-stack framework or frontend application using GoPack templates.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		reader := bufio.NewReader(os.Stdin)

		var framework string
		if len(args) == 0 {
			fmt.Print("Choose framework (Next.js, React, Vue, Svelte, Astro, Tauri): ")
			input, _ := reader.ReadString('\n')
			framework = strings.TrimSpace(input)
		} else {
			framework = args[0]
		}

		if framework == "" {
			framework = "Next.js"
		}

		fmt.Print("Project name: ")
		nameInput, _ := reader.ReadString('\n')
		projectName := strings.TrimSpace(nameInput)
		if projectName == "" {
			projectName = "gopack-app"
		}

		fmt.Print("Use TypeScript? (y/N): ")
		tsInput, _ := reader.ReadString('\n')
		useTS := strings.ToLower(strings.TrimSpace(tsInput)) == "y"

		fmt.Print("Initialize Git? (Y/n): ")
		gitInput, _ := reader.ReadString('\n')
		initGit := strings.ToLower(strings.TrimSpace(gitInput)) != "n"

		fmt.Printf("\nScaffolding %s project in ./%s...\n", framework, projectName)

		// Create files
		err := os.MkdirAll(projectName, 0755)
		if err != nil {
			fmt.Printf("\033[31mError creating directory: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Write a template gopack.json
		manifestContent := fmt.Sprintf(`{
  "name": "%s",
  "version": "1.0.0",
  "scripts": {
    "dev": "gopack run dev",
    "build": "gopack run build"
  },
  "dependencies": {
    "%s": "latest"
  },
  "devDependencies": {}
}`, projectName, strings.ToLower(framework))

		if useTS {
			manifestContent = strings.Replace(manifestContent, `"devDependencies": {}`, `"devDependencies": {\n    "typescript": "latest"\n  }`, 1)
		}

		err = os.WriteFile(filepath.Join(projectName, "gopack.json"), []byte(manifestContent), 0644)
		if err != nil {
			fmt.Printf("\033[31mError writing manifest: %v\033[0m\n", err)
			os.Exit(1)
		}

		// Initialize Git
		if initGit {
			gitCmd := exec.Command("git", "init")
			gitCmd.Dir = projectName
			_ = gitCmd.Run()
			_ = os.WriteFile(filepath.Join(projectName, ".gitignore"), []byte("node_modules/\n.gopack/\n*.lock\n"), 0644)
		}

		fmt.Printf("\033[32mSuccess: Project %s scaffolded successfully!\033[0m\n", projectName)
		fmt.Printf("Run: cd %s && gp install\n", projectName)
	},
}

func init() {
	RootCmd.AddCommand(createCmd)
}
