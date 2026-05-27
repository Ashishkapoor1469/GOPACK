package script

import (
	"bufio"
	"fmt"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

type ScriptRunner struct {
	scripts map[string]string
}

func NewScriptRunner(scripts map[string]string) *ScriptRunner {
	return &ScriptRunner{scripts: scripts}
}

// Run executes a script, invoking its pre- and post- hooks automatically if defined
func (r *ScriptRunner) Run(name string) error {
	// 1. Run pre-hook if exists
	preHook := "pre" + name
	if _, ok := r.scripts[preHook]; ok {
		fmt.Printf("\r\n\033[36m> Executing pre-hook script: %s\033[0m\n", preHook)
		if err := r.executeScriptCmd(preHook); err != nil {
			return err
		}
	}

	// 2. Run target script
	if _, ok := r.scripts[name]; !ok {
		return fmt.Errorf("script not found: %s", name)
	}
	fmt.Printf("\r\n\033[36m> Executing script: %s\033[0m\n", name)
	if err := r.executeScriptCmd(name); err != nil {
		return err
	}

	// 3. Run post-hook if exists
	postHook := "post" + name
	if _, ok := r.scripts[postHook]; ok {
		fmt.Printf("\r\n\033[36m> Executing post-hook script: %s\033[0m\n", postHook)
		if err := r.executeScriptCmd(postHook); err != nil {
			return err
		}
	}

	return nil
}

// RunParallel runs multiple scripts concurrently
func (r *ScriptRunner) RunParallel(names []string) error {
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(names))

	for _, name := range names {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			if err := r.Run(n); err != nil {
				errorsChan <- fmt.Errorf("script %s failed: %w", n, err)
			}
		}(name)
	}

	wg.Wait()
	close(errorsChan)

	var errs []string
	for err := range errorsChan {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return fmt.Errorf("parallel script errors: %s", errs)
	}
	return nil
}

func (r *ScriptRunner) executeScriptCmd(name string) error {
	cmdText := r.scripts[name]

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", cmdText)
	} else {
		cmd = exec.Command("sh", "-c", cmdText)
	}

	// Stream stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	logLine := func(scanner *bufio.Scanner, label string) {
		defer wg.Done()
		for scanner.Scan() {
			timestamp := time.Now().Format("15:04:05.000")
			fmt.Printf("[%s] [%s] %s\n", timestamp, label, scanner.Text())
		}
	}

	go logLine(bufio.NewScanner(stdout), name)
	go logLine(bufio.NewScanner(stderr), name)

	wg.Wait()

	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
