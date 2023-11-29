package utils

import (
	"fmt"
	"os/exec"
	"runtime"
)

func StopService(name string) error {
	path, err := GetMyPath()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		err := Execute("cmd", path, "/c", "sc", "stop", name)
		if err != nil {
			return fmt.Errorf("error stoping service: %v", err)
		}
	case "linux":
		err := Execute("systemctl", path, "stop", name)
		if err != nil {
			return fmt.Errorf("error stoping service: %v", err)
		}
	}
	return nil
}

func UninstallService(name string) error {
	path, err := GetMyPath()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		err := Execute("cmd", path, "/c", "sc", "delete", name)
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
	case "linux":
		err := Execute("systemctl", path, "disable", name)
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
		err = Execute("rm", "/etc/systemd/system/", "/etc/systemd/system/"+name+".service")
		if err != nil {
			return fmt.Errorf("error uninstalling service: %v", err)
		}
	}
	return nil
}

// CheckIfServiceIsInstalled checks if a service is installed
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
