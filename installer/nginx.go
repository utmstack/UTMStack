package main

import (
	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func InstallNginx() error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}

	if err := utils.RunEnvCmd(env, "apt", "update"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt", "install", "-y", "nginx"); err != nil {
		return err
	}

	return nil
}