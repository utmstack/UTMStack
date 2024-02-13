package main

import (
	"path"

	"github.com/utmstack/UTMStack/installer/templates"
	"github.com/utmstack/UTMStack/installer/utils"
)

type Vlan struct {
	Iface string
}

func ConfigureVLAN(mainIface string) error {
	c := Vlan{
		Iface: mainIface,
	}

	err := utils.GenerateConfig(c, templates.Vlan, path.Join("/etc", "netplan", "99-vlan.yaml"))
	if err != nil {
		return err
	}

	if err := utils.RunCmd("netplan", "apply"); err != nil {
		return err
	}

	return nil
}
