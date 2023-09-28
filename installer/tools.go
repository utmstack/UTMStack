package main

import "github.com/utmstack/UTMStack/utils"

func InstallTools() error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}
	err := utils.RunEnvCmd(env, "apt", "install", "-y", "cockpit")
	if err != nil {
		return err
	}

	return nil
}
