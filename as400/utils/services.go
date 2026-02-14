package utils

import (
	"fmt"
	"os"
)

func StopService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "stop", name)
	if err != nil {
		return Logger.ErrorF("error stopping service: %v", err)
	}
	return nil
}

func UninstallService(name string) error {
	path := GetMyPath()
	err := Execute("systemctl", path, "disable", name)
	if err != nil {
		return Logger.ErrorF("error uninstalling service: %v", err)
	}
	err = Execute("rm", "/etc/systemd/system/", "/etc/systemd/system/"+name+".service")
	if err != nil {
		return Logger.ErrorF("error uninstalling service: %v", err)
	}
	return nil
}

func CheckIfServiceIsInstalled(serv string) (bool, error) {
	path := GetMyPath()
	err := Execute("systemctl", path, "status", serv)
	return err == nil, nil
}

func CreateLinuxService(serviceName string, execStart string) error {
	servicePath := "/etc/systemd/system/" + serviceName + ".service"
	if !CheckIfPathExist(servicePath) {
		file, err := os.Create(servicePath)
		if err != nil {
			return Logger.ErrorF("error creating %s file: %v", servicePath, err)
		}
		defer func() { _ = file.Close() }()

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
		return Logger.ErrorF("service %s already exists", serviceName)
	}

	return nil
}
