package utils

import (
	"time"
)

func Uninstall(mode string) error {
	err := RunCmd(mode, "docker-compose", "down")
	if err != nil {
		return err
	}

	// sleep while docker is removing the containers
	time.Sleep(2 * time.Minute)

	// uninstall scanner
	if err := RunCmd(mode, "systemctl", "stop", "utm_scanner"); err != nil {
		return err
	}

	if err := RunCmd(mode, "rm", "/etc/systemd/system/utm_scanner.service"); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "daemon-reload"); err != nil {
		return err
	}

	if err := RunCmd(mode, "rm", "-R", "/opt/scanner/"); err != nil {
		return err
	}

	// uninstall suricata
	if err := RunCmd(mode, "apt", "remove", "-y", "--purge", "suricata"); err != nil {
		return err
	}

	if err := RunCmd(mode, "rm", "-R", "/var/log/suricata"); err != nil {
		return err
	}

	if err := RunCmd(mode, "docker", "volume", "prune", "-f"); err != nil {
		return err
	}

	if err := RunCmd(mode, "docker", "system", "prune", "-f"); err != nil {
		return err
	}

	return nil
}
