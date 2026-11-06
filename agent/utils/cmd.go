package utils

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"unicode/utf8"

	"github.com/threatwinds/go-sdk/entities"
)

// maxConsoleOutput matches the go-sdk entity limit so capped output still validates.
const maxConsoleOutput = entities.MaxEntityAttrLength

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
	outStr := capConsoleOutput(string(out[:]))
	validUtf8Out, _, err := entities.ValidateString(outStr, false)
	if err != nil {
		return outStr + err.Error(), true
	}

	return validUtf8Out, false
}

// capConsoleOutput keeps at most maxConsoleOutput bytes, ending on a rune boundary.
func capConsoleOutput(out string) string {
	if len(out) <= maxConsoleOutput {
		return out
	}
	suffix := fmt.Sprintf("\n... output truncated, showing first %d of %d bytes", maxConsoleOutput, len(out))
	cut := maxConsoleOutput - len(suffix)
	for cut > 0 && !utf8.RuneStart(out[cut]) {
		cut--
	}
	return out[:cut] + fmt.Sprintf("\n... output truncated, showing first %d of %d bytes", cut, len(out))
}
