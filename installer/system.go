package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/utmstack/UTMStack/installer/utils"
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

func PrepareKernel() error {
	mods := []string{
		"8021q",
	}

	f, err := os.OpenFile("/etc/modules", os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	defer f.Close()

	for _, config := range mods {
		if err := utils.RunCmd("modprobe", config); err != nil {
			return fmt.Errorf("failed to load kernel module: %s", config)
		}

		_, err = f.WriteString(config + "\n")
		if err != nil {
			return err
		}
	}

	return nil
}
