package utils

import (
	"fmt"
	"runtime"
	"strings"
)

func CheckIfServiceIsActive(serv string) (bool, error) {
	var errB bool
	var output string
	path := GetMyPath()

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

	serviceStatus := strings.ToLower(strings.TrimSpace(output))
	if runtime.GOOS == "linux" {
		return serviceStatus == "active", nil
	} else if runtime.GOOS == "windows" {
		return strings.Contains(serviceStatus, "running"), nil
	}

	return false, fmt.Errorf("unsupported operating system")
}

func RestartService(serv string) error {
	path := GetMyPath()
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

func StopService(name string) error {
	path := GetMyPath()
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
