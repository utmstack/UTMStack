package utils

func ConfigureUpdater(mode string) error {
	if err := RunCmd(mode, "wget", "http://github.com/AtlasInsideCorp/UTMStackInstaller/releases/latest/download/updater.sh", "-O", "/etc/cron.hourly/updater.sh"); err != nil {
		return err
	}

	if err := RunCmd(mode, "chmod", "755", "/etc/cron.hourly/updater.sh"); err != nil {
		return err
	}

	return nil
}