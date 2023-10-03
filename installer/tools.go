package main

import "github.com/utmstack/UTMStack/installer/utils"

func InstallTools() error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	err := utils.RunEnvCmd(env, "apt", "install", "-y", "cockpit")
	if err != nil {
		return err
	}

	return nil
}
