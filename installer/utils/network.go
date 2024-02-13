package utils

import (
	"net"
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

func GetMainIface(mainIP string) (string, error) {
	var iface string
	ifaces, err := net.Interfaces()
	if err != nil {
		return iface, err
	}

	for _, i := range ifaces {
		al, err := i.Addrs()
		if err != nil {
			return iface, err
		}

		for _, a := range al {
			as := strings.Split(a.String(), "/")[0]

			if as == mainIP {
				iface = i.Name
			}
		}
	}

	return iface, nil
}
