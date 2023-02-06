package utils

import (
	"fmt"
	"time"
)

func InstallOpenVPNMaster(mode string) error {
	if err := RunCmd(mode, "apt-get", "update"); err != nil {
		return err
	}

	if err := RunCmd(mode, "apt-get", "install", "-y", "openvpn", "easy-rsa", "wget"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/ca.crt", "-O", "/etc/openvpn/ca.crt"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/ta.key", "-O", "/etc/openvpn/ta.key"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/dh2048.pem", "-O", "/etc/openvpn/dh2048.pem"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/server.conf", "-O", "/etc/openvpn/server.conf"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/server.crt", "-O", "/etc/openvpn/server.crt"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/server.key", "-O", "/etc/openvpn/server.key"); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "restart", "openvpn"); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return nil
}

func InstallOpenVPNClient(mode, host string) error {
	if err := RunCmd(mode, "apt-get", "update"); err != nil {
		return err
	}

	if err := RunCmd(mode, "apt-get", "install", "-y", "openvpn", "easy-rsa", "wget"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/ca.crt", "-O", "/etc/openvpn/ca.crt"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/ta.key", "-O", "/etc/openvpn/ta.key"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/client.csr", "-O", "/etc/openvpn/client.csr"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/client.conf", "-O", "/etc/openvpn/client.conf"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/client.crt", "-O", "/etc/openvpn/client.crt"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://raw.githubusercontent.com/AtlasInsideCorp/UTMStackInstaller/master/defaults/client.key", "-O", "/etc/openvpn/client.key"); err != nil {
		return err
	}

	sed := fmt.Sprintf("sed -i \"s/IPADDR/%s/g\" /etc/openvpn/client.conf", host)

	if err := RunCmd(mode, "/bin/sh", "-c", sed); err != nil {
		return err
	}

	if err := RunCmd(mode, "systemctl", "restart", "openvpn"); err != nil {
		return err
	}

	time.Sleep(10 * time.Second)

	return nil
}
