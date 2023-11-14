package utils

import (
	"os/exec"
	"runtime"
	"strings"
)

// IsProcessRunning checks if a process is active
func IsProcessRunning(processName string) (bool, error) {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("tasklist")
	} else {
		cmd = exec.Command("ps", "aux")
	}

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	processes := strings.Split(string(output), "\n")
	for _, process := range processes {
		if strings.Contains(process, processName) {
			return true, nil
		}
	}

	return false, nil
}

// StopProcess checks if a process is active, if so it stops it
func StopProcess(processName string) error {
	running, err := IsProcessRunning(processName)
	if err != nil {
		return err
	}

	if running {
		var cmd *exec.Cmd
		if runtime.GOOS == "windows" {
			cmd = exec.Command("taskkill", "/IM", processName, "/F")
		} else {
			cmd = exec.Command("pkill", processName)
		}

		err := cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}
