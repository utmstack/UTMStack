package utils

import "time"

func RemoveServices(services []string) error {
	for _, service := range services {
		if err := RunCmd("docker", "service", "rm", service); err != nil {
			return err
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
