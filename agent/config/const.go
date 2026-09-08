package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/shared/fs"
)

type DataType string

type ProtoPort struct {
	UDP string
	TCP string
}

const (
	SERVICE_UPDATER_NAME = "UTMStackUpdater"
)

var (
	REPLACE_KEY string

	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceLogFile      = filepath.Join(fs.GetExecutablePath(), "logs", "utmstack_agent.log")
	ModulesServName     = "UTMStackModulesLogsCollector"
	WinServName         = "UTMStackWindowsLogsCollector"
	CollectorFileName   = filepath.Join(fs.GetExecutablePath(), "log-collector-config.json")
	UUIDFileName        = filepath.Join(fs.GetExecutablePath(), "uuid.yml")
	ConfigurationFile   = filepath.Join(fs.GetExecutablePath(), "config.yml")
	PortRangeMin        = "1"
	PortRangeMax        = "65535"
	RetentionConfigFile = filepath.Join(fs.GetExecutablePath(), "retention.json")
	LogsDBFile          = filepath.Join(fs.GetExecutablePath(), "logs_process", "logs.db")
	CertPath            = filepath.Join(fs.GetExecutablePath(), "certs", "utm.crt")
	VersionPath         = filepath.Join(fs.GetExecutablePath(), "version.json")
	UpdaterSelfLinux    = "utmstack_updater_self"

	// TLS Configuration for Integrations
	IntegrationCertPath = filepath.Join(fs.GetExecutablePath(), "certs", "integration.crt")
	IntegrationKeyPath  = filepath.Join(fs.GetExecutablePath(), "certs", "integration.key")
	IntegrationCAPath   = filepath.Join(fs.GetExecutablePath(), "certs", "integration-ca.crt")

	DataTypeWindowsAgent        DataType = "wineventlog"
	DataTypeSyslog              DataType = "syslog"
	DataTypeVmware              DataType = "vmware-esxi"
	DataTypeLinuxAgent          DataType = "linux"
	DataTypeEset                DataType = "antivirus-esmc-eset"
	DataTypeKaspersky           DataType = "antivirus-kaspersky"
	DataTypeTraefikModule       DataType = "traefik"
	DataTypeMongodbModule       DataType = "mongodb"
	DataTypeMysqlModule         DataType = "mysql"
	DataTypePostgresqlModule    DataType = "postgresql"
	DataTypeRedisModule         DataType = "redis"
	DataTypeElasticsearchModule DataType = "elasticsearch"
	DataTypeKafkaModule         DataType = "kafka"
	DataTypeKibanaModule        DataType = "kibana"
	DataTypeLogstashModule      DataType = "logstash"
	DataTypeCiscoAsa            DataType = "firewall-cisco-asa"
	DataTypeCiscoMeraki         DataType = "firewall-meraki"
	DataTypeFortinet            DataType = "firewall-fortigate-traffic"
	DataTypePaloalto            DataType = "firewall-paloalto"
	DataTypeMikrotik            DataType = "firewall-mikrotik"
	DataTypeCiscoFirepower      DataType = "firewall-cisco-firepower"
	DataTypeSophosXG            DataType = "firewall-sophos-xg"
	DataTypeCiscoSwitch         DataType = "cisco-switch"
	DataTypeSonicwall           DataType = "firewall-sonicwall"
	DataTypeNatsModule          DataType = "nats"
	DataTypeDeceptivebytes      DataType = "deceptive-bytes"
	DataTypeOsqueryModule       DataType = "osquery"
	DataTypeLinuxAuditdModule   DataType = "auditd"
	DataTypeHaproxyModule       DataType = "haproxy"
	DataTypeNginxModule         DataType = "nginx"
	DataTypeIisModule           DataType = "iis"
	DataTypeApacheModule        DataType = "apache"
	DataTypeSentinelOne         DataType = "antivirus-sentinel-one"
	DataTypeMacOs               DataType = "macos"
	DataTypeGeneric             DataType = "generic"
	DataTypeNetflow             DataType = "netflow"
	DataTypeAix                 DataType = "ibm-aix"
	DataTypePfsense             DataType = "firewall-pfsense"
	DataTypeFortiweb            DataType = "firewall-fortiweb"
	DataTypeSuricata            DataType = "suricata"

	ProtoPorts = map[DataType]ProtoPort{
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
	}

	// FilePaths defines default log file paths for file-based integrations
	FilePaths = map[DataType][]string{
		DataTypeNginxModule:      {"/var/log/nginx/access.log", "/var/log/nginx/error.log"},
		DataTypePostgresqlModule: {"/var/log/postgresql/postgresql-*-main.log"},
	}

	WindowsEventChannelsFile = filepath.Join(fs.GetExecutablePath(), "windows-event-channels.yml")

	// DefaultWindowsEventChannels are always collected. They are defined here and
	// not in WindowsEventChannelsFile so clients can only add to them.
	DefaultWindowsEventChannels = []string{
		"Security",
		"Application",
		"System",
		"Windows Powershell",
		"Microsoft-Windows-Powershell/Operational",
		"ForwardedEvents",
		"Microsoft-Windows-WinLogon/Operational",
		"Microsoft-Windows-Windows Firewall With Advanced Security/Firewall",
		"Microsoft-Windows-Windows Defender/Operational",
	}
)

func ValidateModuleType(typ string) string {
	switch DataType(typ) {
	case DataTypeSyslog, DataTypeVmware, DataTypeEset, DataTypeKaspersky, DataTypeFortinet, DataTypePaloalto,
		DataTypeMikrotik, DataTypeSophosXG, DataTypeSonicwall, DataTypeSentinelOne,
		DataTypeCiscoAsa, DataTypeCiscoFirepower, DataTypeCiscoSwitch, DataTypeCiscoMeraki,
		DataTypeDeceptivebytes, DataTypeAix, DataTypePfsense, DataTypeFortiweb, DataTypeSuricata:
		return "syslog"
	case DataTypeNetflow:
		return "netflow"
	case DataTypeNginxModule, DataTypePostgresqlModule:
		return "file"
	case DataTypeWindowsAgent, DataTypeLinuxAgent, DataTypeTraefikModule, DataTypeMongodbModule, DataTypeMysqlModule,
		DataTypeRedisModule, DataTypeElasticsearchModule, DataTypeKafkaModule, DataTypeKibanaModule, DataTypeLogstashModule, DataTypeNatsModule,
		DataTypeOsqueryModule, DataTypeLinuxAuditdModule, DataTypeHaproxyModule, DataTypeIisModule, DataTypeApacheModule:
		return "beats"
	default:
		return "nil"
	}
}
