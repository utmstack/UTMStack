package utils

import "fmt"

func InstallOpenVPNMaster(mode string) error {
	if err := RunCmd(mode, "apt-get", "install", "-y", "openvpn", "easy-rsa", "wget"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/ca.crt", "-O", "/etc/openvpn/ca.crt"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/ta.key", "-O", "/etc/openvpn/ta.key"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/dh2048.pem", "-O", "/etc/openvpn/dh2048.pem"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/server.conf", "-O", "/etc/openvpn/server.conf"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/server.crt", "-O", "/etc/openvpn/server.crt"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/master/server.key", "-O", "/etc/openvpn/server.key"); err != nil {
		return err
	}
	if err := RunCmd(mode, "systemctl", "restart", "openvpn"); err != nil {
		return err
	}
	return nil
}

func InstallOpenVPNClient(mode, host string) error {
	if err := RunCmd(mode, "apt-get", "install", "-y", "openvpn", "easy-rsa", "wget"); err != nil {
		return err
	}

	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/ca.crt", "-O", "/etc/openvpn/ca.crt"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/ta.key", "-O", "/etc/openvpn/ta.key"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/client.csr", "-O", "/etc/openvpn/client.csr"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/client.conf", "-O", "/etc/openvpn/client.conf"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/client.crt", "-O", "/etc/openvpn/client.crt"); err != nil {
		return err
	}
	if err := RunCmd(mode, "wget", "https://updates.utmstack.com/assets/openvpn/proxy/client.key", "-O", "/etc/openvpn/client.key"); err != nil {
		return err
	}

	sed := fmt.Sprintf("sed -i \"s/IPADDR/%s/g\" /etc/openvpn/client.conf", host)

	if err := RunCmd(mode, "/bin/sh", "-c", sed); err != nil {
		return err
	}
	
	if err := RunCmd(mode, "systemctl", "restart", "openvpn"); err != nil {
		return err
	}
	return nil
}