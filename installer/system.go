package main

import (
	"errors"
	"os"

	"github.com/AtlasInsideCorp/UTMStackInstaller/utils"
)

func PrepareSystem() error {
	sysctl := []string{
		"vm.max_map_count=262144",
		"net.ipv6.conf.all.disable_ipv6=1",
		"net.ipv6.conf.default.disable_ipv6=1",
		"net.ipv4.tcp_keepalive_time=600",
		"net.ipv4.tcp_keepalive_intvl=60",
		"net.ipv4.tcp_keepalive_probes=3",
	}

	f, err := os.OpenFile("/etc/sysctl.conf", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, config := range sysctl {
		if err := utils.RunCmd("sysctl", "-w", config); err != nil {
			return errors.New("failed to set sysctl config")
		}
		_, err = f.WriteString(config + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
