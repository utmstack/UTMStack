package utils

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// CheckIfServiceIsActive checks if a service is active or running
func CheckIfServiceIsActive(serv string) (bool, error) {
	var errB bool
	var output string

	path, err := GetMyPath()
	if err != nil {
		return false, fmt.Errorf("error getting current path: %v", err)
	}

	switch runtime.GOOS {
	case "windows":
		output, errB = ExecuteWithResult("sc", path, "query", serv)
	case "linux":
		output, errB = ExecuteWithResult("systemctl", path, "is-active", serv)
	default:
		return false, fmt.Errorf("unknown operating system")
	}

	if errB {
		return false, nil
	}

	serviceStatus := strings.ToLower(strings.TrimSpace(string(output)))
	if runtime.GOOS == "linux" {
		return serviceStatus == "active", nil
	} else if runtime.GOOS == "windows" {
		return strings.Contains(serviceStatus, "running"), nil
	}

	return false, fmt.Errorf("unsupported operating system")
}

// RestartService restarts a service
func RestartService(serv string) error {
	path, err := GetMyPath()
	if err != nil {
		return fmt.Errorf("error getting current path: %v", err)
	}

	isRunning, err := CheckIfServiceIsActive(serv)
	if err != nil {
		return fmt.Errorf("error checking if %s service is active: %v", serv, err)
	}

	switch runtime.GOOS {
	case "windows":
		if isRunning {
			err := Execute("sc", path, "stop", serv)
			if err != nil {
				return fmt.Errorf("error stopping service: %v", err)
			}
		}
		err := Execute("sc", path, "start", serv)
		if err != nil {
			return fmt.Errorf("error starting service: %v", err)
		}

	case "linux":
		if isRunning {
			err := Execute("systemctl", path, "restart", serv)
			if err != nil {
				return fmt.Errorf("error restarting service: %v", err)
			}
		} else {
			err := Execute("systemctl", path, "start", serv)
			if err != nil {
				return fmt.Errorf("error starting service: %v", err)
			}
		}
	}
	return nil
}

// StopService stops a service
func StopService(name string) error {
	path, err := GetMyPath()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		err := Execute("sc", path, "stop", name)
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

// UninstallService uninstalls a service
func UninstallService(name string) error {
	path, err := GetMyPath()
	if err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		err := Execute("sc", path, "delete", name)
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
