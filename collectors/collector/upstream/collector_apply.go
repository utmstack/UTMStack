package upstream

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/utmstack/UTMStack/collectors/forwarder/collector"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

type groupConfig struct {
	enabled         string
	proto           string
	port            string
	tls             string
	auth            string
	path            string
	signatureHeader string
}

func parseGroupConfigurations(entries []*CollectorGroupConfigurations) groupConfig {
	var gc groupConfig
	for _, e := range entries {
		if e == nil {
			continue
		}
		switch e.GetConfKey() {
		case "enabled":
			gc.enabled = e.GetConfValue()
		case "proto":
			gc.proto = e.GetConfValue()
		case "port":
			gc.port = e.GetConfValue()
		case "tls":
			gc.tls = e.GetConfValue()
		case "auth":
			gc.auth = e.GetConfValue()
		case "path":
			gc.path = e.GetConfValue()
		case "signature_header":
			gc.signatureHeader = e.GetConfValue()
		}
	}
	return gc
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func rejectResult(msg string) *ConfigKnowledge {
	return &ConfigKnowledge{Accepted: "false", ErrorMessage: msg}
}

func acceptResult() *ConfigKnowledge {
	return &ConfigKnowledge{Accepted: "true"}
}

func applyCollectorConfigGroup(group *CollectorConfigGroup) *ConfigKnowledge {
	if group == nil {
		return rejectResult("nil collector config group")
	}

	dataType := group.GetGroupName()
	if dataType == "" {
		return rejectResult("missing data type (group_name)")
	}

	conf := parseGroupConfigurations(group.GetConfigurations())

	if conf.enabled == "" {
		return rejectResult(fmt.Sprintf("integration %q: missing required conf_key %q", dataType, "enabled"))
	}
	if conf.proto == "" {
		return rejectResult(fmt.Sprintf("integration %q: missing required conf_key %q", dataType, "proto"))
	}

	enabled := conf.enabled == "true"

	if enabled {
		if _, err := collector.ResolveOrCreateDataType(dataType, conf.proto, conf.port); err != nil {
			return rejectResult(fmt.Sprintf("integration %q: %v", dataType, err))
		}
	}

	switch conf.proto {
	case "tcp", "udp", "tls":
		return applyNetworkIntegration(dataType, conf, enabled)
	case "http", "https":
		return applyHTTPIntegration(dataType, conf, enabled)
	default:
		return rejectResult(fmt.Sprintf("integration %q: invalid protocol %q", dataType, conf.proto))
	}
}

func applyNetworkIntegration(dataType string, conf groupConfig, enabled bool) *ConfigKnowledge {
	useTLS := conf.proto == "tls" || conf.tls == "true"
	networkProto := conf.proto
	if useTLS {
		networkProto = "tcp"
	}

	if !enabled {
		if _, known := config.ResolveDataType(dataType); !known {
			return acceptResult()
		}
	}

	if _, err := collector.ChangeIntegrationStatus(dataType, networkProto, enabled, useTLS); err != nil {
		return rejectResult(fmt.Sprintf("integration %q: %v", dataType, err))
	}
	return acceptResult()
}

func applyHTTPIntegration(dataType string, conf groupConfig, enabled bool) *ConfigKnowledge {
	if !enabled {
		if err := collector.DisableHTTPIntegration(dataType, conf.proto); err != nil {
			return rejectResult(fmt.Sprintf("integration %q: %v", dataType, err))
		}
		return acceptResult()
	}

	if conf.port == "" {
		return rejectResult(fmt.Sprintf("integration %q: missing required conf_key %q for proto %q", dataType, "port", conf.proto))
	}

	// Defaults for a generic HTTP integration.
	defaultPath, defaultBind, defaultAuth, defaultSigHeader := "/logs", "127.0.0.1", "", "X-Hub-Signature-256"

	if spec, ok := config.HTTPPorts[config.DataType(dataType)]; ok {
		defaultPath, defaultBind = spec.Path, spec.Bind
		if spec.Auth != "" {
			defaultAuth = spec.Auth
		}
		if spec.SignatureHeader != "" {
			defaultSigHeader = spec.SignatureHeader
		}
	}

	opts := collector.HTTPIntegrationOptions{
		Proto:           conf.proto,
		Port:            conf.port,
		Path:            defaultString(conf.path, defaultPath),
		Bind:            defaultBind,
		Auth:            defaultString(conf.auth, defaultAuth),
		SignatureHeader: defaultString(conf.signatureHeader, defaultSigHeader),
	}
	generatedSecret, err := collector.EnableHTTPIntegration(dataType, opts)
	if err != nil {
		return rejectResult(fmt.Sprintf("integration %q: %v", dataType, err))
	}
	result := acceptResult()
	result.GeneratedSecret = generatedSecret
	return result
}

const (
	tlsCertConfKeyAction  = "action"
	tlsCertConfKeyPayload = "payload_enc"
	tlsCertConfKeyNonce   = "nonce"

	tlsCertActionApply  = "apply"
	tlsCertActionStatus = "status"
)

type tlsCertEnvelope struct {
	CertPem string `json:"certPem"`
	KeyPem  string `json:"keyPem"`
	CaPem   string `json:"caPem"`
}

func applyTLSCertGroup(group *CollectorConfigGroup) *ConfigKnowledge {
	if group == nil {
		return rejectResult("nil collector config group")
	}

	confs := group.GetConfigurations()
	action := tlsConfValue(confs, tlsCertConfKeyAction)

	switch action {
	case tlsCertActionApply:
		return applyTLSCertPush(confs)
	case tlsCertActionStatus:
		return tlsStatusResult()
	default:
		return rejectResult(fmt.Sprintf("%s: unknown action %q", config.ReservedTLSCertsGroup, action))
	}
}

func tlsConfValue(confs []*CollectorGroupConfigurations, key string) string {
	for _, c := range confs {
		if c.GetConfKey() == key {
			return c.GetConfValue()
		}
	}
	return ""
}

func applyTLSCertPush(confs []*CollectorGroupConfigurations) *ConfigKnowledge {
	payloadB64 := tlsConfValue(confs, tlsCertConfKeyPayload)
	nonceB64 := tlsConfValue(confs, tlsCertConfKeyNonce)
	if payloadB64 == "" || nonceB64 == "" {
		return rejectResult(fmt.Sprintf("%s: missing %s/%s", config.ReservedTLSCertsGroup, tlsCertConfKeyPayload, tlsCertConfKeyNonce))
	}

	cnf, err := config.GetCurrentConfig()
	if err != nil {
		return rejectResult(fmt.Sprintf("%s: failed to load local config: %v", config.ReservedTLSCertsGroup, err))
	}

	key := utils.DeriveTLSCertKey(strconv.FormatUint(uint64(cnf.CollectorID), 10), cnf.CollectorKey)
	plaintext, err := utils.OpenTLSCertEnvelope(key, payloadB64, nonceB64)
	if err != nil {
		return rejectResult(fmt.Sprintf("%s: %v", config.ReservedTLSCertsGroup, err))
	}

	var envelope tlsCertEnvelope
	if err := json.Unmarshal(plaintext, &envelope); err != nil {
		return rejectResult(fmt.Sprintf("%s: invalid decrypted envelope: %v", config.ReservedTLSCertsGroup, err))
	}
	if envelope.CertPem == "" || envelope.KeyPem == "" {
		return rejectResult(fmt.Sprintf("%s: decrypted envelope is missing cert or key", config.ReservedTLSCertsGroup))
	}

	if _, err := tls.X509KeyPair([]byte(envelope.CertPem), []byte(envelope.KeyPem)); err != nil {
		return rejectResult(fmt.Sprintf("%s: certificate and private key do not match: %v", config.ReservedTLSCertsGroup, err))
	}

	if err := writeIntegrationCerts(envelope); err != nil {
		return rejectResult(fmt.Sprintf("%s: %v", config.ReservedTLSCertsGroup, err))
	}
	return acceptResult()
}

func writeIntegrationCerts(envelope tlsCertEnvelope) error {
	stagingDir, err := os.MkdirTemp("", "utmstack-tls-certs-")
	if err != nil {
		return fmt.Errorf("failed to create staging directory: %w", err)
	}
	defer os.RemoveAll(stagingDir)

	src := utils.CertificateFiles{
		CertPath: filepath.Join(stagingDir, "cert.pem"),
		KeyPath:  filepath.Join(stagingDir, "key.pem"),
		CAPath:   filepath.Join(stagingDir, "ca.pem"),
	}
	if err := os.WriteFile(src.CertPath, []byte(envelope.CertPem), 0600); err != nil {
		return fmt.Errorf("failed to stage certificate: %w", err)
	}
	if err := os.WriteFile(src.KeyPath, []byte(envelope.KeyPem), 0600); err != nil {
		return fmt.Errorf("failed to stage private key: %w", err)
	}
	if envelope.CaPem != "" {
		if err := os.WriteFile(src.CAPath, []byte(envelope.CaPem), 0600); err != nil {
			return fmt.Errorf("failed to stage CA certificate: %w", err)
		}
	}

	dest := utils.CertificateFiles{
		CertPath: config.IntegrationCertPath,
		KeyPath:  config.IntegrationKeyPath,
		CAPath:   config.IntegrationCAPath,
	}
	return utils.LoadUserCertificatesWithStruct(src, dest)
}

func tlsStatusResult() *ConfigKnowledge {
	tlsStatus := utils.GetTLSStatus(config.IntegrationCertPath, config.IntegrationKeyPath, config.IntegrationCAPath)
	statusJSON, err := json.Marshal(tlsStatus)
	if err != nil {
		return rejectResult(fmt.Sprintf("%s: failed to marshal tls status: %v", config.ReservedTLSCertsGroup, err))
	}
	return &ConfigKnowledge{Accepted: "true", StatusPayload: string(statusJSON)}
}
