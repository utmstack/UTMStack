package utils

import (
	"errors"
	"net"
	"strings"
)

func GetIPAddress() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
			if ip := localAddr.IP.To4(); ip != nil && !ip.IsLoopback() {
				return ip.String(), nil
			}
		}
	}

	return "", errors.New("failed to get IP address")
}

func CleanIPAddresses(input string) string {
	var clean []string

	addresses := strings.Split(input, ",")
	for _, addr := range addresses {
		addr = strings.TrimSpace(addr)

		ip := net.ParseIP(addr)
		if ip == nil {
			if ipNet, _, err := net.ParseCIDR(addr); err == nil {
				ip = ipNet
			}
		}

		if ip == nil {
			continue
		}

		if ip.IsLoopback() {
			continue
		}

		clean = append(clean, ip.String())
	}

	return strings.Join(clean, ",")
}
