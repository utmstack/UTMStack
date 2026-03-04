// Package schema defines shared configuration types for collector packages.
// This package exists to avoid circular imports between collector, syslog, and netflow.
package schema

import (
	"encoding/json"
	"os"
	"strconv"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/shared/fs"
)

// Port represents a TCP or UDP port configuration.
type Port struct {
	IsListen   bool   `json:"enabled"`
	Port       string `json:"value"`
	TLSEnabled bool   `json:"tls_enabled,omitempty"`
}

// Integration represents a collector integration configuration (syslog/netflow).
type Integration struct {
	TCP Port `json:"tcp_port,omitempty"`
	UDP Port `json:"udp_port,omitempty"`
}

// FileIntegration represents a file-based log collector configuration.
type FileIntegration struct {
	Enabled bool     `json:"enabled"`
	Paths   []string `json:"paths"`
}

// CollectorConfig represents the full collector configuration file.
type CollectorConfig struct {
	Integrations     map[string]Integration     `json:"integrations"`
	FileIntegrations map[string]FileIntegration `json:"file_integrations,omitempty"`
}

// ReadCollectorConfig reads the collector configuration from disk.
func ReadCollectorConfig() (CollectorConfig, error) {
	cnf := CollectorConfig{}
	err := fs.ReadJSON(config.CollectorFileName, &cnf)
	if err != nil {
		return cnf, err
	}
	return cnf, nil
}

// WriteCollectorConfig writes the collector configuration to disk.
func WriteCollectorConfig(cnf *CollectorConfig) error {
	data, err := json.MarshalIndent(cnf, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(config.CollectorFileName, data, 0644)
}

// ValidatePortChange returns true if the port is within the valid range.
func ValidatePortChange(newPort string) bool {
	port, err := strconv.Atoi(newPort)
	if err != nil {
		return false
	}
	min, _ := strconv.Atoi(config.PortRangeMin)
	max, _ := strconv.Atoi(config.PortRangeMax)
	return port >= min && port <= max
}
