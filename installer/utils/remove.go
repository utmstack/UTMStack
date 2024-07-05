package utils

import (
	"time"

	"github.com/threatwinds/logger"
)

func RemoveServices(services []string) error {
	for _, service := range services {
		if err := RunCmd("docker", "service", "rm", service); err != nil {
			if !logger.Is(err, "not found"){
				return err
			}
		}
	}

	if err := RunCmd("systemctl", "restart", "docker"); err != nil {
		return err
	}

	time.Sleep(60 * time.Second)

	if err := RunCmd("docker", "system", "prune", "-f"); err != nil {
		return err
	}

	return nil
}
