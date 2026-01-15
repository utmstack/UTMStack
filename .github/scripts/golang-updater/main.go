package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	modFile, err := os.Open("go.mod")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening go.mod: %v\n", err)
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "Error reading go.mod: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command("go", "list", "-u", "-m", "-json", "all")
	output, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing go list: %v\n", err)
		os.Exit(1)
	}

	decoder := json.NewDecoder(bytes.NewReader(output))
	var toUpdate []string

	for decoder.More() {
		var mod Module
		if err := decoder.Decode(&mod); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON output: %v\n", err)
			os.Exit(1)
		}
		if mod.Update != nil && explicitModules[mod.Path] {
			toUpdate = append(toUpdate, fmt.Sprintf("%s@%s", mod.Path, mod.Update.Version))
		}
	}

	if len(toUpdate) == 0 {
		fmt.Println("✅ There are no updates available for the explicitly required modules.")
		return
	}

	for _, mod := range toUpdate {
		fmt.Printf("🔄 Updating %s\n", mod)
		cmd := exec.Command("go", "get", mod)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Error updating %s: %v\n", mod, err)
		}
	}

	fmt.Println("🧹 Executing go mod tidy...")
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing go mod tidy: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Dependencies updated successfully.")
}
