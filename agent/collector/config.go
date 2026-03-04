package collector

import (
	"fmt"
	"net"

	"github.com/utmstack/UTMStack/agent/collector/schema"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

// SyncCollectorConfig ensures the collector config file exists and is synchronized
// with the current ProtoPorts and FilePaths. It handles both initial creation and version upgrades:
// - If config doesn't exist → creates it with all integrations from ProtoPorts and FilePaths
// - If config exists → adds new integrations, removes obsolete ones
func SyncCollectorConfig() error {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		// Config file doesn't exist, create it
		integrations := make(map[string]schema.Integration)
		for logTyp, ports := range config.ProtoPorts {
			integrations[string(logTyp)] = schema.Integration{
				TCP: schema.Port{IsListen: false, Port: ports.TCP},
				UDP: schema.Port{IsListen: false, Port: ports.UDP},
			}
		}

		fileIntegrations := make(map[string]schema.FileIntegration)
		for logTyp, paths := range config.FilePaths {
			fileIntegrations[string(logTyp)] = schema.FileIntegration{
				Enabled: false,
				Paths:   paths,
			}
		}

		newConfig := &schema.CollectorConfig{
			Integrations:     integrations,
			FileIntegrations: fileIntegrations,
		}
		return schema.WriteCollectorConfig(newConfig)
	}

	modified := false

	// Add new integrations from ProtoPorts
	for logTyp, ports := range config.ProtoPorts {
		key := string(logTyp)
		if _, exists := cnf.Integrations[key]; !exists {
			cnf.Integrations[key] = schema.Integration{
				TCP: schema.Port{IsListen: false, Port: ports.TCP},
				UDP: schema.Port{IsListen: false, Port: ports.UDP},
			}
			modified = true
			utils.Logger.Info("Added new integration to config: %s", key)
		}
	}

	// Remove integrations that no longer exist in ProtoPorts
	for key := range cnf.Integrations {
		if _, exists := config.ProtoPorts[config.DataType(key)]; !exists {
			delete(cnf.Integrations, key)
			modified = true
			utils.Logger.Info("Removed obsolete integration from config: %s", key)
		}
	}

	// Initialize FileIntegrations map if nil
	if cnf.FileIntegrations == nil {
		cnf.FileIntegrations = make(map[string]schema.FileIntegration)
	}

	// Add new file integrations from FilePaths
	for logTyp, paths := range config.FilePaths {
		key := string(logTyp)
		if _, exists := cnf.FileIntegrations[key]; !exists {
			cnf.FileIntegrations[key] = schema.FileIntegration{
				Enabled: false,
				Paths:   paths,
			}
			modified = true
			utils.Logger.Info("Added new file integration to config: %s", key)
		}
	}

	// Remove file integrations that no longer exist in FilePaths
	for key := range cnf.FileIntegrations {
		if _, exists := config.FilePaths[config.DataType(key)]; !exists {
			delete(cnf.FileIntegrations, key)
			modified = true
			utils.Logger.Info("Removed obsolete file integration from config: %s", key)
		}
	}

	if modified {
		return schema.WriteCollectorConfig(&cnf)
	}
	return nil
}

// ConfigureFirstTime is an alias for SyncCollectorConfig for backward compatibility.
// Deprecated: Use SyncCollectorConfig instead.
func ConfigureFirstTime() error {
	return SyncCollectorConfig()
}

// ChangeIntegrationStatus enables or disables an integration.
func ChangeIntegrationStatus(logTyp string, proto string, isEnabled bool, tlsOptions ...bool) (string, error) {
	var port string
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		port = integration.TCP.Port
	case "udp":
		port = integration.UDP.Port
	default:
		return "", fmt.Errorf("invalid protocol: %s", proto)
	}

	// When enabling, validate port is available
	if isEnabled {
		if conflicting := CheckPortConflict(port, proto, &cnf, logTyp); conflicting != "" {
			return "", fmt.Errorf("port %s is already in use by integration '%s'", port, conflicting)
		}
		if !IsPortBindable(port, proto) {
			return "", fmt.Errorf("port %s is already in use by another process", port)
		}
	}

	switch proto {
	case "tcp":
		integration.TCP.IsListen = isEnabled

		if len(tlsOptions) > 0 && isEnabled {
			if tlsOptions[0] {
				if !fs.Exists(config.IntegrationCertPath) || !fs.Exists(config.IntegrationKeyPath) {
					return "", fmt.Errorf("TLS certificates not found. Please load certificates first")
				}
				integration.TCP.TLSEnabled = true
			} else {
				integration.TCP.TLSEnabled = false
			}
		}

		if !isEnabled {
			integration.TCP.TLSEnabled = false
		}

	case "udp":
		integration.UDP.IsListen = isEnabled

		if len(tlsOptions) > 0 && tlsOptions[0] {
			return "", fmt.Errorf("TLS is not supported for UDP protocol. Use TCP for TLS connections")
		}
	}

	cnf.Integrations[logTyp] = integration
	return port, schema.WriteCollectorConfig(&cnf)
}

// ChangePortResult contains the result of a port change operation.
type ChangePortResult struct {
	OldPort string
	Warning string
}

// ChangePort changes the port for an integration.
// Returns the old port and a warning if another integration has the same port configured.
func ChangePort(logTyp string, proto string, port string) (ChangePortResult, error) {
	result := ChangePortResult{}
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return result, fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return result, fmt.Errorf("invalid integration: %s", logTyp)
	}

	if !schema.ValidatePortChange(port) {
		return result, fmt.Errorf("port %s is out of valid range %s-%s", port, config.PortRangeMin, config.PortRangeMax)
	}

	// Check if port is in use by an enabled integration (error)
	if conflicting := CheckPortConflict(port, proto, &cnf, logTyp); conflicting != "" {
		return result, fmt.Errorf("port %s is already in use by integration '%s'", port, conflicting)
	}

	// Check if port is configured by another integration (warning)
	if configured := CheckPortConfigured(port, proto, &cnf, logTyp); configured != "" {
		result.Warning = fmt.Sprintf("port %s is also configured for integration '%s' (currently disabled)", port, configured)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		result.OldPort = integration.TCP.Port
		integration.TCP.Port = port
	case "udp":
		result.OldPort = integration.UDP.Port
		integration.UDP.Port = port
	default:
		return result, fmt.Errorf("invalid protocol: %s", proto)
	}

	cnf.Integrations[logTyp] = integration
	return result, schema.WriteCollectorConfig(&cnf)
}

// CheckPortConflict checks if a port is in use by another enabled integration.
// Returns the integration name if there's a conflict, empty string otherwise.
func CheckPortConflict(port string, proto string, cnf *schema.CollectorConfig, currentIntegration string) string {
	for integration, integrationConfig := range cnf.Integrations {
		if integration == currentIntegration {
			continue
		}
		if proto == "tcp" && integrationConfig.TCP.Port == port && integrationConfig.TCP.IsListen {
			return integration
		}
		if proto == "udp" && integrationConfig.UDP.Port == port && integrationConfig.UDP.IsListen {
			return integration
		}
	}
	return ""
}

// CheckPortConfigured checks if a port is configured (but not necessarily enabled) by another integration.
// Returns the integration name if configured, empty string otherwise.
func CheckPortConfigured(port string, proto string, cnf *schema.CollectorConfig, currentIntegration string) string {
	for integration, integrationConfig := range cnf.Integrations {
		if integration == currentIntegration {
			continue
		}
		if proto == "tcp" && integrationConfig.TCP.Port == port {
			return integration
		}
		if proto == "udp" && integrationConfig.UDP.Port == port {
			return integration
		}
	}
	return ""
}

// IsPortBindable checks if the system can bind to a port.
func IsPortBindable(port string, proto string) bool {
	addr := ":" + port
	switch proto {
	case "tcp":
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return false
		}
		listener.Close()
		return true
	case "udp":
		conn, err := net.ListenPacket("udp", addr)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	default:
		return false
	}
}

// EnableTLSForIntegration enables TLS for an integration.
func EnableTLSForIntegration(logTyp string, proto string) (string, error) {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid == "nil" {
		return "", fmt.Errorf("invalid integration: %s", logTyp)
	}

	integration := cnf.Integrations[logTyp]
	var port string

	switch proto {
	case "tcp":
		if integration.TCP.Port == "" {
			return "", fmt.Errorf("TCP port not configured for %s", logTyp)
		}
		port = integration.TCP.Port
		integration.TCP.TLSEnabled = true
	case "udp":
		return "", fmt.Errorf("TLS not supported for UDP protocol")
	default:
		return "", fmt.Errorf("invalid protocol: %s", proto)
	}

	cnf.Integrations[logTyp] = integration
	return port, schema.WriteCollectorConfig(&cnf)
}

// DisableTLSForIntegration disables TLS for an integration.
func DisableTLSForIntegration(logTyp string, proto string) error {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector config: %v", err)
	}

	integration := cnf.Integrations[logTyp]
	switch proto {
	case "tcp":
		integration.TCP.TLSEnabled = false
	case "udp":
		return fmt.Errorf("TLS not supported for UDP protocol")
	default:
		return fmt.Errorf("invalid protocol: %s", proto)
	}

	cnf.Integrations[logTyp] = integration
	return schema.WriteCollectorConfig(&cnf)
}

// ChangeFileIntegrationStatus enables or disables a file-based integration.
func ChangeFileIntegrationStatus(logTyp string, isEnabled bool) ([]string, error) {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid != "file" {
		return nil, fmt.Errorf("invalid file integration: %s", logTyp)
	}

	integration, exists := cnf.FileIntegrations[logTyp]
	if !exists {
		return nil, fmt.Errorf("file integration not found: %s", logTyp)
	}

	integration.Enabled = isEnabled
	cnf.FileIntegrations[logTyp] = integration
	return integration.Paths, schema.WriteCollectorConfig(&cnf)
}

// ChangeFilePathsResult contains the result of a path change operation.
type ChangeFilePathsResult struct {
	OldPaths []string
}

// ChangeFilePaths changes the paths for a file-based integration.
func ChangeFilePaths(logTyp string, paths []string) (ChangeFilePathsResult, error) {
	result := ChangeFilePathsResult{}
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return result, fmt.Errorf("error reading collector config: %v", err)
	}

	if valid := config.ValidateModuleType(logTyp); valid != "file" {
		return result, fmt.Errorf("invalid file integration: %s", logTyp)
	}

	integration, exists := cnf.FileIntegrations[logTyp]
	if !exists {
		return result, fmt.Errorf("file integration not found: %s", logTyp)
	}

	if len(paths) == 0 {
		return result, fmt.Errorf("at least one path is required")
	}

	result.OldPaths = integration.Paths
	integration.Paths = paths
	cnf.FileIntegrations[logTyp] = integration
	return result, schema.WriteCollectorConfig(&cnf)
}

// GetFileIntegrationPaths returns the configured paths for a file integration.
func GetFileIntegrationPaths(logTyp string) ([]string, error) {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading collector config: %v", err)
	}

	integration, exists := cnf.FileIntegrations[logTyp]
	if !exists {
		return nil, fmt.Errorf("file integration not found: %s", logTyp)
	}

	return integration.Paths, nil
}
