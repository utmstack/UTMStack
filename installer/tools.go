package main

import "github.com/utmstack/UTMStack/installer/utils"

func InstallTools(distro string) error {
	switch distro {
	case "ubuntu":
		env := []string{"DEBIAN_FRONTEND=noninteractive"}
		if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "cockpit"); err != nil {
			return err
		}
	case "redhat":
		env := []string{"DNF_YUM_AUTO_YES=1"}
		if err := utils.RunEnvCmd(env, "dnf", "install", "-y", "cockpit"); err != nil {
			return err
		}
	}

	return nil
}
