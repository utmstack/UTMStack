package utils

import (
	"fmt"
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

func GetMainIPInAirGapMode() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var candidateIPs []string
	var ethernetIPs []string

	for _, i := range ifaces {
		// Skip down interfaces and loopback interfaces
		if i.Flags&net.FlagUp == 0 || i.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Skip Docker and virtual interfaces (common patterns)
		ifaceName := strings.ToLower(i.Name)
		if strings.HasPrefix(ifaceName, "docker") ||
			strings.HasPrefix(ifaceName, "br-") ||
			strings.HasPrefix(ifaceName, "veth") ||
			strings.HasPrefix(ifaceName, "virbr") {
			continue
		}

		addrs, err := i.Addrs()
		if err != nil {
			continue // Skip interfaces that fail to return addresses
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			// We are only interested in IPv4 addresses that are not loopback
			if ip == nil || ip.IsLoopback() || ip.To4() == nil {
				continue
			}

			// Skip link-local addresses (169.254.x.x)
			if ip.To4()[0] == 169 && ip.To4()[1] == 254 {
				continue
			}

			ipStr := ip.String()

			// Prioritize Ethernet interfaces
			if strings.HasPrefix(ifaceName, "eth") ||
				strings.HasPrefix(ifaceName, "ens") ||
				strings.HasPrefix(ifaceName, "enp") {
				ethernetIPs = append(ethernetIPs, ipStr)
			} else {
				candidateIPs = append(candidateIPs, ipStr)
			}
		}
	}

	// Return first Ethernet IP if available
	if len(ethernetIPs) > 0 {
		return ethernetIPs[0], nil
	}

	// Otherwise return first candidate IP
	if len(candidateIPs) > 0 {
		return candidateIPs[0], nil
	}

	return "", fmt.Errorf("could not find a suitable local IP address in offline mode")
}

// ResolveMainIP resolves the host's primary IP, falling back to local interface
// discovery or a previously persisted value from /root/utmstack.yml when the
// network is still coming up at boot (e.g. air-gap or delayed DHCP).
func ResolveMainIP(online bool, persisted string, getOnlineIP, getAirGapIP func() (string, error)) (string, error) {
	if online {
		if ip, err := getOnlineIP(); err == nil && ip != "" {
			return ip, nil
		}
		if ip, err := getAirGapIP(); err == nil && ip != "" {
			return ip, nil
		}
		if persisted != "" {
			return persisted, nil
		}
		return "", fmt.Errorf("could not determine main IP online or via local interfaces")
	}

	ip, err := getAirGapIP()
	if err == nil && ip != "" {
		return ip, nil
	}
	if persisted != "" {
		return persisted, nil
	}
	return "", err
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
