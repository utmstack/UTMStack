package collector

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/collectors/forwarder/collector/schema"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	listeners "github.com/utmstack/UTMStack/collectors/forwarder/listeners"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

func SyncCollectorConfig() error {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		integrations := make(map[string]schema.Integration)
		for logTyp, pp := range config.ProtoPorts {
			integrations[string(logTyp)] = schema.Integration{
				TCP: schema.Port{IsListen: false, Port: pp.TCP},
				UDP: schema.Port{IsListen: false, Port: pp.UDP},
			}
		}

		fileIntegrations := make(map[string]schema.FileIntegration)
		for logTyp, path := range config.FilePaths {
			fileIntegrations[string(logTyp)] = schema.FileIntegration{
				Enabled: false,
				Paths:   []string{path},
			}
		}

		newConfig := &schema.CollectorConfig{
			Integrations:     integrations,
			FileIntegrations: fileIntegrations,
		}
		return schema.WriteCollectorConfig(newConfig)
	}

	modified := false

	for logTyp, pp := range config.ProtoPorts {
		key := string(logTyp)
		if _, exists := cnf.Integrations[key]; !exists {
			cnf.Integrations[key] = schema.Integration{
				TCP: schema.Port{IsListen: false, Port: pp.TCP},
				UDP: schema.Port{IsListen: false, Port: pp.UDP},
			}
			modified = true
			utils.Logger.Info("Added new integration to config: %s", key)
		}
	}

	for key := range cnf.Integrations {
		if _, exists := config.ProtoPorts[config.DataType(key)]; !exists {
			integration := cnf.Integrations[key]
			if integration.HTTP != nil || integration.HTTPS != nil {
				continue
			}
			delete(cnf.Integrations, key)
			modified = true
			utils.Logger.Info("Removed obsolete integration from config: %s", key)
		}
	}

	if cnf.FileIntegrations == nil {
		cnf.FileIntegrations = make(map[string]schema.FileIntegration)
	}

	for logTyp, path := range config.FilePaths {
		key := string(logTyp)
		if _, exists := cnf.FileIntegrations[key]; !exists {
			cnf.FileIntegrations[key] = schema.FileIntegration{
				Enabled: false,
				Paths:   []string{path},
			}
			modified = true
			utils.Logger.Info("Added new file integration to config: %s", key)
		}
	}

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

func ConfigureFirstTime() error {
	return SyncCollectorConfig()
}

func ChangeIntegrationStatus(logTyp string, proto string, isEnabled bool, tlsOptions ...bool) (string, error) {
	var port string
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return "", fmt.Errorf("error reading collector config: %v", err)
	}

	if _, ok := config.ResolveDataType(logTyp); !ok {
		return "", fmt.Errorf("unknown integration: %s", logTyp)
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

type ChangePortResult struct {
	OldPort string
	Warning string
}

func ChangePort(logTyp string, proto string, port string) (ChangePortResult, error) {
	result := ChangePortResult{}
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return result, fmt.Errorf("error reading collector config: %v", err)
	}

	if _, ok := config.ResolveDataType(logTyp); !ok {
		return result, fmt.Errorf("unknown integration: %s", logTyp)
	}

	if !schema.ValidatePortChange(port) {
		return result, fmt.Errorf("port %s is out of valid range %d-%d", port, config.PortRangeMin, config.PortRangeMax)
	}

	if conflicting := CheckPortConflict(port, proto, &cnf, logTyp); conflicting != "" {
		return result, fmt.Errorf("port %s is already in use by integration '%s'", port, conflicting)
	}

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

func ChangeFileIntegrationStatus(logTyp string, isEnabled bool) ([]string, error) {
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return nil, fmt.Errorf("error reading collector config: %v", err)
	}

	if _, ok := config.FilePaths[config.DataType(logTyp)]; !ok {
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

type ChangeFilePathsResult struct {
	OldPaths []string
}

func ChangeFilePaths(logTyp string, paths []string) (ChangeFilePathsResult, error) {
	result := ChangeFilePathsResult{}
	cnf, err := schema.ReadCollectorConfig()
	if err != nil {
		return result, fmt.Errorf("error reading collector config: %v", err)
	}

	if _, ok := config.FilePaths[config.DataType(logTyp)]; !ok {
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

type HTTPIntegrationOptions struct {
	Proto           string // "http" or "https"
	Port            string
	Path            string // URL path, default "/logs"
	Bind            string // bind address, default "127.0.0.1"
	Auth            string // "" | "bearer" | "hmac"
	SignatureHeader string // e.g. "X-Hub-Signature-256"
}

func EnableHTTPIntegration(logTyp string, opts HTTPIntegrationOptions) error {
	cfg, err := schema.ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector config: %v", err)
	}

	if cfg.Integrations == nil {
		cfg.Integrations = make(map[string]schema.Integration)
	}

	if opts.Auth == "bearer" || opts.Auth == "hmac" {
		tp := filepath.Join(config.HTTPTokenDir, "integration-http-"+logTyp+".token")
		if _, statErr := os.Stat(tp); statErr != nil {
			// Generate token when file doesn't exist or can't be read
			token, genErr := listeners.GenerateTokenFile(tp)
			if genErr != nil {
				return fmt.Errorf("generate token: %w", genErr)
			}
			fmt.Printf("\nToken for %s: %s\n", logTyp, token)
			fmt.Printf("Store this token securely — it won't be shown again.\n\n")
		}
	}

	integration := cfg.Integrations[logTyp]
	httpPort := &schema.HTTPPort{
		Enabled:         true,
		Port:            opts.Port,
		Path:            opts.Path,
		Bind:            opts.Bind,
		Auth:            opts.Auth,
		SignatureHeader: opts.SignatureHeader,
	}

	if opts.Proto == "https" {
		integration.HTTPS = httpPort
	} else {
		integration.HTTP = httpPort
	}
	cfg.Integrations[logTyp] = integration

	return schema.WriteCollectorConfig(&cfg)
}

func DisableHTTPIntegration(logTyp string, proto string) error {
	cfg, err := schema.ReadCollectorConfig()
	if err != nil {
		return fmt.Errorf("error reading collector config: %v", err)
	}

	integration, ok := cfg.Integrations[logTyp]
	if !ok {
		return nil // nothing to disable
	}

	if proto == "https" {
		integration.HTTPS = nil
	} else {
		integration.HTTP = nil
	}
	cfg.Integrations[logTyp] = integration

	return schema.WriteCollectorConfig(&cfg)
}
