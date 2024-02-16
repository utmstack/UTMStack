package utils

import (
	"os"
	"os/exec"
	"path/filepath"

)

func RunEnvCmd(env []string, command string, arg ...string) error {
	cmd := exec.Command(command, arg...)
	cmd.Env = append(os.Environ(), env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func RunCmd(command string, arg ...string) error {
	return RunEnvCmd([]string{}, command, arg...)
}

func MakeDir(mode os.FileMode, arg ...string) string {
	path := ""
	for _, folder := range arg {
		path = filepath.Join(path, folder)
		os.MkdirAll(path, mode)
		os.Chmod(path, mode)
	}
	return path
}

func Remove(path string) error{
	return os.Remove(path)
}