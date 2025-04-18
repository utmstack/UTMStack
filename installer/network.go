package main

import (
	"fmt"
	"log"
	"os"
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
	folder := path.Join("/etc", "netplan")
	files, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		bytes, err := os.ReadFile(path.Join(folder, file.Name()))
		if err != nil {
			return "", err
		}

		if strings.Contains(fmt.Sprintf("%s", bytes), "NetworkManager") {
			return "NetworkManager", nil
		}
	}

	return "networkd", nil
}

func ConfigureVLAN(mainIface string, distro string) error {
	c := Vlan{
		Iface: mainIface,
	}

	switch distro {
	case "ubuntu":
		renderer, err := checkRenderer()
		if err != nil {
			return err
		}
		c.Renderer = renderer

		log.Println("Generating vlan config")
		err = utils.GenerateConfig(c, templates.VlanUbuntu, path.Join("/etc", "netplan", "99-vlan.yaml"))
		if err != nil {
			return err
		}

		log.Println("Applying vlan config")
		if err := utils.RunCmd("netplan", "apply"); err != nil {
			return err
		}

	case "redhat":
		err := utils.GenerateConfig(c, templates.VlanRedHat, "/etc/sysconfig/network-scripts/ifcfg-vlan10")
		if err != nil {
			return err
		}

		if err := utils.RunCmd("modprobe", "8021q"); err != nil {
			return err
		}

		if err := os.WriteFile("/etc/modules-load.d/8021q.conf", []byte("8021q\n"), 0644); err != nil {
			return err
		}

		if err := utils.RunCmd("systemctl", "restart", "NetworkManager"); err != nil {
			return err
		}

	}
	return nil
}

func InstallVlan(distro string) error {
	switch distro {
	case "ubuntu":
		env := []string{"DEBIAN_FRONTEND=noninteractive"}

		if err := utils.RunEnvCmd(env, "apt-get", "update", "-y"); err != nil {
			return err
		}

		if err := utils.RunEnvCmd(env, "apt-get", "install", "-y", "vlan"); err != nil {
			return err
		}

	case "redhat":
		env := []string{"DNF_YUM_AUTO_YES=1"}

		if err := utils.RunEnvCmd(env, "dnf", "update", "-y"); err != nil {
			return err
		}
	}

	return nil
}
