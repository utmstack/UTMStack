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
	time.Sleep(20 * time.Second)

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
