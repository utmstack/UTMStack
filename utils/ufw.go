package utils

import (
	"errors"
	"path"
)

func ConfigureFirewall(mode string, c Config) error {
	if err := RunCmd(mode, "apt-get", "install", "-y", "ufw"); err != nil {
		return err
	}

	if err := GenerateFromTemplate(c, ufw, path.Join("/", "etc", "ufw", "after.rules")); err != nil {
		return errors.New("failed to generate after.rules file: " + err.Error())
	}

	return nil
}
