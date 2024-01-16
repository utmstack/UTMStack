package utils

import (
	"errors"
	"os/exec"
	"unicode/utf8"
)

// ExecuteWithResult executes a command and returns the output
func ExecuteWithResult(c string, dir string, arg ...string) (string, bool) {
	cmd := exec.Command(c, arg...)

	cmd.Dir = dir
	if errors.Is(cmd.Err, exec.ErrDot) {
		cmd.Err = nil
	}

	out, err := cmd.Output()
	if err != nil {
		return string(out[:]) + err.Error(), true
	}

	validUtf8Out := cleanString(string(out[:]))

	return validUtf8Out, false
}

// Execute executes a command without returning the output
func Execute(c string, dir string, arg ...string) error {
	cmd := exec.Command(c, arg...)
	cmd.Dir = dir

	return cmd.Run()
}

// cleanString cleans a string from invalid utf8 characters
func cleanString(s string) string {
	v := make([]rune, 0, len(s))
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				v = append(v, '?')
				continue
			}
		}
		v = append(v, r)
	}
	return string(v)
}
