package utils

import (
	"errors"
	"os/exec"

	"github.com/threatwinds/go-sdk/entities"
)

func ExecuteWithResult(c string, dir string, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	out, err := cmd.Output()
	if err != nil {
		return string(out) + err.Error(), true
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
