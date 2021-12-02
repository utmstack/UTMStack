package utils

func InstallScanner(mode string) error {
	if err := RunCmd(mode, "apt", "update"); err != nil {
		return err
	}
	
	if err := RunCmd(mode, "apt", "install", "-y", "python3", "python3-pip", "zip", "unzip"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "-O", "scanner.zip", "https://updates.utmstack.com/assets/scanner.zip"); err != nil {
		return err
	}

	if err := RunCmd(mode, "unzip", "scanner.zip"); err != nil {
		return err
	}

	if err := RunCmd(mode, "mv", "UTMStackHostScanner/", "/opt/scanner"); err != nil {
		return err
	}

	if err := RunCmd(mode, "mv", "/opt/scanner/utm_scanner.service", "/etc/systemd/system/"); err != nil {
		return err
	}

	if err := RunCmd(mode, "chmod", "777", "/opt/scanner/run.sh"); err != nil {
		return err
	}

	if err := RunCmd(mode, "pip3", "install", "-r", "/opt/scanner/requirements.txt"); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "enable", "utm_scanner"); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "start", "utm_scanner"); err != nil {
		return err
	}

	return nil
}
