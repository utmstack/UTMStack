// Package exec provides utilities for executing external commands.
package exec

import (
	"bytes"
	"fmt"
	"os/exec"
)

// Run executes a command with the given arguments in the specified working directory.
// Returns an error if the command fails.
func Run(command, workDir string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("command failed: %s: %w", stderr.String(), err)
		}
		return fmt.Errorf("command failed: %w", err)
	}
	return nil
}

// RunWithOutput executes a command and returns its combined stdout and stderr output.
func RunWithOutput(command, workDir string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}
	return string(output), nil
}
