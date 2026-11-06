package utils

import (
	"bytes"
	"errors"
	"os/exec"

	"github.com/threatwinds/go-sdk/entities"
)

func ExecuteWithResult(c string, dir string, arg ...string) (string, bool) {
	return runWithResult(c, dir, false, arg...)
}

// ExecuteWithResultRaw passes the command line to the shell unescaped on
// Windows, because Go's default quoting adds backslashes that cmd.exe rejects.
func ExecuteWithResultRaw(c string, dir string, arg ...string) (string, bool) {
	return runWithResult(c, dir, true, arg...)
}

func runWithResult(c string, dir string, rawWindows bool, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)
	if rawWindows {
		setRawCmdLine(cmd, c, arg)
	}

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		return string(out) + stderr.String() + err.Error(), true
	}

	if string(out[:]) == "" {
		return "Command executed successfully but no output", false
	}
	validUtf8Out, _, err := entities.ValidateString(string(out[:]), false)
	if err != nil {
		return string(out) + err.Error(), true
	}

	return validUtf8Out, false
}
