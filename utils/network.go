package utils

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strings"
)

func GetMainIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

func GetIfaceIP(iface string) (string, error) {
	ief, err := net.InterfaceByName(iface)

	if err != nil {
		return "", err
	}

	addrs, err := ief.Addrs()
	if err != nil {
		return "", err
	}

	cidr := addrs[0].String()
	
	addr := strings.Split(cidr, "/")

	if len(addr) == 0 {
		return "", fmt.Errorf("cannot get %s address", iface)
	}

	return addr[0], nil
}

func GetMainIface(mode string) (string, error) {
	err := RunCmd(mode, "apt", "update")
	if err != nil {
		return "", err
	}
	err = RunCmd(mode, "apt", "install", "-y", "wget", "net-tools")
	if err != nil {
		return "", err
	}
	cmd := exec.Command("/bin/sh", "-c", "route | grep '^default' | grep -o '[^ ]*$'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	err = cmd.Start()
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	s := buf.String()
	return s[:len(s)-1], nil
}
