package utils

import (
	"bytes"
	"log"
	"net"
	"os/exec"
)

func GetMainIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

func GetMainIface(mode string) string {
	if err := RunCmd(mode, "apt", "update"); err != nil {
		log.Fatal(err)
	}

	if err := RunCmd(mode, "apt", "install", "-y", "wget", "net-tools"); err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command("/bin/sh", "-c", "route | grep '^default' | grep -o '[^ ]*$'")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(stdout)
	s := buf.String()
	return s[:len(s)-1]
}
