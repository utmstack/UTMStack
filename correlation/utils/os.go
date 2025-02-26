package utils

import (
	"os"
	"os/exec"
)

func RunCommand(command ...string) error {
	cmd := exec.Command(command[0], command[1:]...)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
