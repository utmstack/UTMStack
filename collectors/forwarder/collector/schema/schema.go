// Package schema defines shared configuration types for collector packages.
// This package exists to avoid circular imports between collector, syslog, and netflow.
package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/shared/fs"
)

var configMu sync.Mutex

func LockCollectorConfig() {
	configMu.Lock()
}

func UnlockCollectorConfig() {
	configMu.Unlock()
}

type Port struct {
	IsListen   bool   `json:"enabled"`
	Port       string `json:"value"`
	TLSEnabled bool   `json:"tls_enabled,omitempty"`
}

type HTTPPort struct {
	Enabled         bool   `json:"enabled"`
	Port            string `json:"value"`
	Path            string `json:"path"`                 // default "/logs"
	Bind            string `json:"bind"`                 // "127.0.0.1" default
	Auth            string `json:"auth,omitempty"`       // "" | "bearer" | "hmac"
	SignatureHeader string `json:"sig_header,omitempty"` // e.g. "X-Hub-Signature-256"
}

type Integration struct {
	TCP   Port      `json:"tcp_port,omitempty"`
	UDP   Port      `json:"udp_port,omitempty"`
	HTTP  *HTTPPort `json:"http_port,omitempty"`
	HTTPS *HTTPPort `json:"https_port,omitempty"`
}

type FileIntegration struct {
	Enabled bool     `json:"enabled"`
	Paths   []string `json:"paths"`
}

type CollectorConfig struct {
	Integrations     map[string]Integration     `json:"integrations"`
	FileIntegrations map[string]FileIntegration `json:"file_integrations,omitempty"`
}

func ReadCollectorConfig() (CollectorConfig, error) {
	cnf := CollectorConfig{}
	err := fs.ReadJSON(config.CollectorFileName, &cnf)
	if err != nil {
		return cnf, err
	}
	return cnf, nil
}

func WriteCollectorConfig(cnf *CollectorConfig) error {
	data, err := json.MarshalIndent(cnf, "", "    ")
	if err != nil {
		return err
	}

	dest := config.CollectorFileName
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return fmt.Errorf("error writing temp collector config: %w", err)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return fmt.Errorf("error saving collector config: %w", err)
	}
	return nil
}

func ValidatePortChange(newPort string) bool {
	port, err := strconv.Atoi(newPort)
	if err != nil {
		return false
	}
	min := config.PortRangeMin
	max := config.PortRangeMax
	return port >= min && port <= max
}
