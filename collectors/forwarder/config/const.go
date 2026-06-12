package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/shared/fs"
)

// REPLACE_KEY is injected at build time via -ldflags.
var REPLACE_KEY string

var (
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"

	CollectorLogFile  = filepath.Join(fs.GetExecutablePath(), "logs", "utmstack_forwarder.log")
	UUIDFileName      = filepath.Join(fs.GetExecutablePath(), "collector-uuid.yml")
	ConfigurationFile = filepath.Join(fs.GetExecutablePath(), "collector-config.yml")
	CollectorFileName = filepath.Join(fs.GetExecutablePath(), "log-collector-config.json")
	CatalogFileName   = filepath.Join(fs.GetExecutablePath(), "log-collector-catalog.json")
	CatalogLockFile   = filepath.Join(fs.GetExecutablePath(), "log-collector-catalog.lock")
	HTTPTokenDir      = filepath.Join(fs.GetExecutablePath(), "certs")

	IntegrationCertPath = filepath.Join(fs.GetExecutablePath(), "certs", "integration.crt")
	IntegrationKeyPath  = filepath.Join(fs.GetExecutablePath(), "certs", "integration.key")
	IntegrationCAPath   = filepath.Join(fs.GetExecutablePath(), "certs", "integration_ca.crt")

	ServiceName      = "UTMStackForwarder"
	ServiceArguments = []string{"run"}
)

const (
	PortRangeMin = 1
	PortRangeMax = 65535
)

// DataType represents the type of log data being collected from an external integration.
type DataType string

const (
	DataTypeGeneric          DataType = "generic"
	DataTypeSyslog           DataType = "syslog"
	DataTypeVmware           DataType = "vmware-esxi"
	DataTypeEset             DataType = "antivirus-esmc-eset"
	DataTypeKaspersky        DataType = "antivirus-kaspersky"
	DataTypeCiscoAsa         DataType = "firewall-cisco-asa"
	DataTypeCiscoFirepower   DataType = "firewall-cisco-firepower"
	DataTypeCiscoSwitch      DataType = "cisco-switch"
	DataTypeCiscoMeraki      DataType = "firewall-meraki"
	DataTypeFortinet         DataType = "firewall-fortigate-traffic"
	DataTypePaloalto         DataType = "firewall-paloalto"
	DataTypeMikrotik         DataType = "firewall-mikrotik"
	DataTypeSophosXG         DataType = "firewall-sophos-xg"
	DataTypeSonicwall        DataType = "firewall-sonicwall"
	DataTypeDeceptivebytes   DataType = "deceptive-bytes"
	DataTypeSentinelOne      DataType = "antivirus-sentinel-one"
	DataTypeAix              DataType = "ibm-aix"
	DataTypePfsense          DataType = "firewall-pfsense"
	DataTypeFortiweb         DataType = "firewall-fortiweb"
	DataTypeSuricata         DataType = "suricata"
	DataTypeNetflow          DataType = "netflow"
	DataTypeNginxModule      DataType = "nginx"
	DataTypePostgresqlModule DataType = "postgresql"
	DataTypeGithub           DataType = "github"
	DataTypeOracle           DataType = "oracle"
)

type ProtoPort struct {
	UDP string
	TCP string
}

var ProtoPorts = map[DataType]ProtoPort{
	DataTypeSyslog:         {UDP: "7014", TCP: "7014"},
	DataTypeVmware:         {UDP: "7002", TCP: "7002"},
	DataTypeEset:           {UDP: "7003", TCP: "7003"},
	DataTypeKaspersky:      {UDP: "7004", TCP: "7004"},
	DataTypeCiscoAsa:       {UDP: "514", TCP: "1470"},
	DataTypeCiscoFirepower: {UDP: "514", TCP: "1470"},
	DataTypeCiscoSwitch:    {UDP: "514", TCP: "1470"},
	DataTypeCiscoMeraki:    {UDP: "514", TCP: "1470"},
	DataTypeFortinet:       {UDP: "7005", TCP: "7005"},
	DataTypePaloalto:       {UDP: "7006", TCP: "7006"},
	DataTypeMikrotik:       {UDP: "7007", TCP: "7007"},
	DataTypeSophosXG:       {UDP: "7008", TCP: "7008"},
	DataTypeSonicwall:      {UDP: "7009", TCP: "7009"},
	DataTypeDeceptivebytes: {UDP: "7010", TCP: "7010"},
	DataTypeSentinelOne:    {UDP: "7012", TCP: "7012"},
	DataTypeAix:            {UDP: "7016", TCP: "7016"},
	DataTypePfsense:        {UDP: "7017", TCP: "7017"},
	DataTypeFortiweb:       {UDP: "7018", TCP: "7018"},
	DataTypeSuricata:       {UDP: "7019", TCP: "7019"},
	DataTypeNetflow:        {UDP: "2055", TCP: ""},
	DataTypeOracle:         {UDP: "7021", TCP: "7021"},
}

var FilePaths = map[DataType]string{
	DataTypeNginxModule:      "/var/log/nginx/",
	DataTypePostgresqlModule: "/var/log/postgresql/",
}

type HTTPSpec struct {
	Proto           string // "http" | "https"
	Port            string
	Path            string // URL path the sender posts to
	Bind            string // "0.0.0.0" for externally-reachable webhooks
	Auth            string // "" | "bearer" | "hmac"
	SignatureHeader string // header carrying the HMAC signature (hmac auth)
}

var HTTPPorts = map[DataType]HTTPSpec{
	DataTypeGithub: {
		Proto:           "https",
		Port:            "7020",
		Path:            "/github",
		Bind:            "0.0.0.0",
		Auth:            "hmac",
		SignatureHeader: "X-Hub-Signature-256",
	},
}
