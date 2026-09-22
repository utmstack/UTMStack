package collector

import (
	"fmt"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
)

func ResolveOrCreateDataType(name, proto, port string) (config.DataTypeEntry, error) {
	entry, known := config.ResolveDataType(name)
	if known {
		return entry, nil
	}

	if port == "" {
		return config.DataTypeEntry{}, fmt.Errorf("port is required for new integration %q: it has no built-in default", name)
	}

	kind := "syslog"
	switch proto {
	case "http", "https":
		kind = proto
	case "netflow":
		kind = "netflow"
	}

	udp, tcp := "", ""
	switch proto {
	case "udp":
		udp = port
	case "tcp", "tls":
		tcp = port
	}

	if err := config.AddUserDataType(name, kind, udp, tcp); err != nil {
		return config.DataTypeEntry{}, fmt.Errorf("failed to create data type %q: %w", name, err)
	}

	if err := seedIntegrationPorts(name, udp, tcp); err != nil {
		return config.DataTypeEntry{}, fmt.Errorf("failed to seed runtime config for data type %q: %w", name, err)
	}

	entry, _ = config.ResolveDataType(name)
	return entry, nil
}
