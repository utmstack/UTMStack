package utils

import "path"

func ConfigureFirewall(mode string) error {
	if err := RunCmd(mode, "apt-get", "install", "-y", "ufw"); err != nil {
		return err
	}

	if err := WriteToFile(path.Join("/", "etc", "ufw", "after.rules"), ufw); err != nil {
		return err
	}

	return nil
}
