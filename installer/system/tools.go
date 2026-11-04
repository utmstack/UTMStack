package system

import (
	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

func InstallTools(distro string) error {
	switch distro {
	case config.RequiredDistroUbuntu, config.RequiredDistroDebian:
		env := []string{"DEBIAN_FRONTEND=noninteractive"}
		if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "cockpit"); err != nil {
			return err
		}
	case config.RequiredDistroRHEL:
		env := []string{"DNF_YUM_AUTO_YES=1"}
		if err := utils.RunEnvCmd(env, "dnf", "install", "-y", "cockpit"); err != nil {
			return err
		}
	}

	return nil
}
