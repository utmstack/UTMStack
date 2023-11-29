package utils

import (
	"os/exec"
)

func Execute(c string, dir string, arg ...string) error {
	cmd := exec.Command(c, arg...)
	cmd.Dir = dir

	return cmd.Run()
}
