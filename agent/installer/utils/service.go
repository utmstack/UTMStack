package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("sc", "query", serv)
	case "linux":
		cmd = exec.Command("systemctl", "status", serv)
	default:
		return false, fmt.Errorf("operative system unknown")
	}

	if err := cmd.Run(); err != nil {
		return false, nil
	}
	return true, nil
}
