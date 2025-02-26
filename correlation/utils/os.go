package utils

import (
	"io"
	"log"
	"os/exec"
)

func RunCommand(command ...string) error {
	cmd := exec.Command(command[0], command[1:]...)
	output, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	bytesRead, err := io.ReadAll(output)
	if err != nil {
		return err
	}

	err = output.Close()
	if err != nil {
		return err
	}

	log.Println(string(bytesRead))

	return nil
}
