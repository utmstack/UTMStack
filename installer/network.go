package main

import (
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

		if strings.Contains(string(bytes), "NetworkManager") {
			return "NetworkManager", nil
		}
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

	log.Println("Generating vlan config")
	err = utils.GenerateConfig(c, templates.Vlan, path.Join("/etc", "netplan", "99-vlan.yaml"))
	if err != nil {
		return err
	}

	log.Println("Applying vlan config")
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
