package utils

import (
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/shared/exec"
	"github.com/utmstack/UTMStack/shared/fs"
)

func CheckIfServiceIsActive(serv string) (bool, error) {
	path := fs.GetExecutablePath()
	output, errB := ExecuteWithResult("systemctl", path, "is-active", serv)
	if errB {
		return false, nil
	}
	return output == "active", nil
}

func StartService(name string) error {
	path := fs.GetExecutablePath()
	if err := exec.Run("systemctl", path, "start", name); err != nil {
		return fmt.Errorf("error starting service: %v", err)
	}
	return nil
}

func StopService(name string) error {
	path := fs.GetExecutablePath()
	if err := exec.Run("systemctl", path, "stop", name); err != nil {
		return fmt.Errorf("error stopping service: %v", err)
	}
	return nil
}

func UninstallService(name string) error {
	path := fs.GetExecutablePath()
	if err := exec.Run("systemctl", path, "disable", name); err != nil {
		return fmt.Errorf("error disabling service: %v", err)
	}
	if err := exec.Run("rm", path, "/etc/systemd/system/"+name+".service"); err != nil {
		return fmt.Errorf("error removing service file: %v", err)
	}
	return nil
}

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	path := fs.GetExecutablePath()
	err := exec.Run("systemctl", path, "status", serv)
	return err == nil, nil
}

func CreateLinuxService(serviceName string, execStart string) error {
	servicePath := "/etc/systemd/system/" + serviceName + ".service"
	if fs.Exists(servicePath) {
		return fmt.Errorf("service %s already exists", serviceName)
	}

	file, err := os.Create(servicePath)
	if err != nil {
		return fmt.Errorf("error creating %s: %v", servicePath, err)
	}
	defer file.Close()

	content := fmt.Sprintf(`[Unit]
Description=%s
After=network.target

[Service]
ExecStart=%s
Restart=always

[Install]
WantedBy=multi-user.target
`, serviceName, execStart)

	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return file.Sync()
}
