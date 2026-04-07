package utils

import "net"

func GetHostAliases(hostname string) ([]string, error) {
	var aliases []string
	addresses, err := net.LookupHost(hostname)
	if err != nil {
		return nil, err
	}

	for _, address := range addresses {
		newAliases, err := net.LookupAddr(address)
		if err != nil {
			return nil, err
		}
		aliases = append(aliases, newAliases...)
	}

	return aliases, nil
}
