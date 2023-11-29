package utils

import (
	"fmt"
	"os"
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
	path, err := GetMyPath()
	if err != nil {
		return false, err
	}
	switch runtime.GOOS {
	case "windows":
		err = Execute("sc", path, "query", serv)
	case "linux":
		err = Execute("systemctl", path, "status", serv)
	default:
		return false, fmt.Errorf("operative system unknown")
	}

	if err != nil {
		return false, nil
	}
	return true, nil
}

func CreateLinuxService(serviceName string, execStart string) error {
	servicePath := "/etc/systemd/system/" + serviceName + ".service"
	if !CheckIfPathExist(servicePath) {
		file, err := os.Create(servicePath)
		if err != nil {
			return fmt.Errorf("error creating %s file: %v", servicePath, err)
		}
		defer file.Close()

		serviceContent := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
ExecStart=%s
Restart=always

[Install]
WantedBy=multi-user.target
`, serviceName, execStart)

		_, err = file.WriteString(serviceContent)
		if err != nil {
			return err
		}

		err = file.Sync()
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("service %s already exists", serviceName)
	}

	return nil
}
