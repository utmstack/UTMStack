package utils

import (
	"bytes"
	"net"
	"os/exec"
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
