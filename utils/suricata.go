package utils

import "fmt"

func InstallSuricata(mode string, iface string) error {
	
	if err := RunCmd(mode, "apt", "update"); err != nil {
		return err
	}

	if err := RunCmd(mode, "apt", "install", "-y", "software-properties-common"); err != nil {
		return err
	}

	if err := RunCmd(mode, "add-apt-repository", "-y", "ppa:oisf/suricata-stable"); err != nil {
		return err
	}

	env := []string{
		"DEBIAN_FRONTEND=noninteractive",
	}

	if err := RunEnvCmd(mode, env, "apt", "install", "-y", "suricata"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "-O", "/etc/suricata/suricata.yaml", "https://updates.utmstack.com/assets/suricata.yaml"); err != nil {
		return err
	}

	sed := fmt.Sprintf("sed -i \"s/SCANNER_IFACE/%s/g\" /etc/suricata/suricata.yaml", iface)

	if err := RunCmd(mode, "/bin/sh", "-c", sed); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "restart", "suricata"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "update-sources"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "malsilo/win-malware"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "tgreen/hunting"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "etnetera/aggressive"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "sslbl/ja3-fingerprints"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "sslbl/ssl-fp-blacklist"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "ptresearch/attackdetection"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update", "enable-source", "oisf/trafficid"); err != nil {
		return err
	}

	if err := RunCmd(mode, "suricata-update"); err != nil {
		return err
	}

	return nil
}
