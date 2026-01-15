package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Module struct {
	Path    string
	Version string
	Update  *ModuleUpdate
}

type ModuleUpdate struct {
	Version string
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	var checkOnly bool
	var discover bool
	var targetPath string

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--check":
			checkOnly = true
		case "--update":
			checkOnly = false
		case "--discover":
			discover = true
		case "--help", "-h":
			printUsage()
			os.Exit(0)
		default:
			if !strings.HasPrefix(arg, "--") {
				targetPath = arg
			}
		}
	}

	// Validate arguments
	if !discover && targetPath == "" {
		fmt.Fprintf(os.Stderr, "Error: must specify a path or use --discover\n")
		printUsage()
		os.Exit(1)
	}

	if discover && targetPath != "" {
		fmt.Fprintf(os.Stderr, "Error: cannot use both --discover and a specific path\n")
		printUsage()
		os.Exit(1)
	}

	var projects []string
	var err error

	if discover {
		projects, err = discoverProjects(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error discovering projects: %v\n", err)
			os.Exit(1)
		}
		if len(projects) == 0 {
			fmt.Println("No Go projects found.")
			os.Exit(0)
		}
		fmt.Printf("🔍 Discovered %d Go projects\n\n", len(projects))
	} else {
		// Verify the path exists and has a go.mod
		goModPath := filepath.Join(targetPath, "go.mod")
		if _, err := os.Stat(goModPath); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: no go.mod found in %s\n", targetPath)
			os.Exit(1)
		}
		projects = []string{targetPath}
	}

	hasUpdates := false
	allUpdates := make(map[string][]Module)

	for _, project := range projects {
		updates, err := checkProject(project)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error checking %s: %v\n", project, err)
			os.Exit(1)
		}
		if len(updates) > 0 {
			hasUpdates = true
			allUpdates[project] = updates
		}
	}

	if !hasUpdates {
		fmt.Println("✅ All dependencies are up to date.")
		return
	}

	// Print summary of updates needed
	fmt.Println("📦 Dependencies with updates available:")
	for project, updates := range allUpdates {
		fmt.Printf("\n  📁 %s:\n", project)
		for _, mod := range updates {
			fmt.Printf("     - %s: %s → %s\n", mod.Path, mod.Version, mod.Update.Version)
		}
	}

	if checkOnly {
		fmt.Println("\n❌ Please update dependencies before merging.")
		os.Exit(1)
	}

	// Update mode - apply updates
	fmt.Println("\n🔄 Updating dependencies...")
	for project, updates := range allUpdates {
		fmt.Printf("\n  📁 %s:\n", project)
		if err := updateProject(project, updates); err != nil {
			fmt.Fprintf(os.Stderr, "Error updating %s: %v\n", project, err)
			os.Exit(1)
		}
	}

	fmt.Println("\n✅ All dependencies updated successfully.")
}

func printUsage() {
	fmt.Println(`Usage: golang-updater [--check|--update] [--discover|<path>]

Modes:
  --check     Check for outdated dependencies (exit 1 if found)
  --update    Update outdated dependencies (default)

Target:
  --discover  Discover all Go projects from current directory
  <path>      Path to a specific Go project

Examples:
  golang-updater --check ./installer
  golang-updater --update ./installer
  golang-updater --check --discover
  golang-updater --update --discover`)
}

func discoverProjects(root string) ([]string, error) {
	var projects []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and common non-project directories
		if info.IsDir() {
			name := info.Name()
			// Don't skip the root directory itself
			if path != root && (strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules") {
				return filepath.SkipDir
			}
		}

		if info.Name() == "go.mod" {
			dir := filepath.Dir(path)
			projects = append(projects, dir)
		}

		return nil
	})

	return projects, err
}

func checkProject(projectPath string) ([]Module, error) {
	goModPath := filepath.Join(projectPath, "go.mod")
	modFile, err := os.Open(goModPath)
	if err != nil {
		return nil, fmt.Errorf("error opening go.mod: %w", err)
	}
	defer modFile.Close()

	explicitModules := make(map[string]bool)
	scanner := bufio.NewScanner(modFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "require") || strings.HasPrefix(line, ")") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 1 && !strings.HasPrefix(fields[0], "//") {
			explicitModules[fields[0]] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading go.mod: %w", err)
	}

	cmd := exec.Command("go", "list", "-u", "-m", "-json", "all")
	cmd.Dir = projectPath
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("error executing go list: %w", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	var toUpdate []Module

	for decoder.More() {
		var mod Module
		if err := decoder.Decode(&mod); err != nil {
			return nil, fmt.Errorf("error parsing JSON output: %w", err)
		}
		if mod.Update != nil && explicitModules[mod.Path] {
			toUpdate = append(toUpdate, mod)
		}
	}

	return toUpdate, nil
}

func updateProject(projectPath string, updates []Module) error {
	for _, mod := range updates {
		updateStr := fmt.Sprintf("%s@%s", mod.Path, mod.Update.Version)
		fmt.Printf("     🔄 Updating %s\n", updateStr)
		cmd := exec.Command("go", "get", updateStr)
		cmd.Dir = projectPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error updating %s: %w", updateStr, err)
		}
	}

	fmt.Printf("     🧹 Running go mod tidy...\n")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error running go mod tidy: %w", err)
	}

	return nil
}
