package main

import (
	"os/exec"
	"path"
	"strings"

	"github.com/utmstack/UTMStack/installer/templates"
	"github.com/utmstack/UTMStack/installer/utils"
)

type Vlan struct {
	Renderer string
	Iface    string
}

func checkRenderer() (string, error) {
	out, err := exec.Command("bash", "-c", "grep renderer /etc/netplan/*.yaml").Output()
	if err != nil {
		return "", err
	}

	if strings.Contains(string(out), "NetworkManager") {
		return "NetworkManager", nil
	} else if strings.Contains(string(out), "networkd") {
		return "networkd", nil
	}

	return "networkd", nil
}

func ConfigureVLAN(mainIface string) error {
	renderer, err := checkRenderer()
	if err != nil {
		return err
	}

	c := Vlan{
		Renderer: renderer,
		Iface:    mainIface,
	}

	err = utils.GenerateConfig(c, templates.Vlan, path.Join("/etc", "netplan", "99-vlan.yaml"))
	if err != nil {
		return err
	}

	if err := utils.RunCmd("netplan", "apply"); err != nil {
		return err
	}

	return nil
}

func InstallVlan() error {
	env := []string{"DEBIAN_FRONTEND=noninteractive"}

	if err := utils.RunEnvCmd(env, "apt-get", "update"); err != nil {
		return err
	}

	if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "vlan"); err != nil {
		return err
	}

	return nil
}
