package main

import "github.com/AtlasInsideCorp/UTMStackInstaller/utils"

func InstallTools() error {
	err := utils.RunCmd("apt", "install", "-y", "cockpit")
	if err != nil {
		return err
	}

	return nil
}
