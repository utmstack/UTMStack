package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/threatwinds/validations"
)

func ExecuteWithResult(c string, dir string, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out[:]) + err.Error(), true
	}

	validUtf8Out, _, err := validations.ValidateString(string(out[:]), false)
	if err != nil {
		return string(out[:]) + err.Error(), true
	}

	return validUtf8Out, false
}

func Execute(c string, dir string, arg ...string) error {
	cmd := exec.Command(c, arg...)
	if dir != "" {
		cmd.Dir = dir
	}

	return cmd.Run()
}

func ExecuteWithContext(ctx context.Context, c string, dir string, arg ...string) error {
	cmd := exec.CommandContext(ctx, c, arg...)
	if dir != "" {
		cmd.Dir = dir
	}

	return cmd.Run()
}

func GetCmdAndArgFromString(cmd string) (string, []string) {
	parts := strings.Fields(cmd)
	base := parts[0]
	args := parts[1:]

	return base, args
}

func CheckErrorsInOutput(output string) error {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "ERROR") {
			return errors.New(line)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading standard input: %v", err)
	}
	return nil
}
