package utils

import (
	"errors"
	"net"
)

func GetIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("failed to get IP address")
}

func IsPortUsed(proto string, port string) bool {
	conn, err := net.Listen(proto, ":"+port)
	if err != nil {
		return true
	}
	conn.Close()
	return false
}
