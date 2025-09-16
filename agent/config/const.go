package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/utils"
)

type DataType string

type ProtoPort struct {
	UDP string
	TCP string
}

var (
	REPLACE_KEY string

	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceLogFile      = filepath.Join(utils.GetMyPath(), "logs", "utmstack_agent.log")
	ModulesServName     = "UTMStackModulesLogsCollector"
	WinServName         = "UTMStackWindowsLogsCollector"
	CollectorFileName   = filepath.Join(utils.GetMyPath(), "log-collector-config.json")
	UUIDFileName        = filepath.Join(utils.GetMyPath(), "uuid.yml")
	ConfigurationFile   = filepath.Join(utils.GetMyPath(), "config.yml")
	PortRangeMin        = "7000"
	PortRangeMax        = "9000"
	RetentionConfigFile = filepath.Join(utils.GetMyPath(), "retention.json")
	LogsDBFile          = filepath.Join(utils.GetMyPath(), "logs_process", "logs.db")
	CertPath            = filepath.Join(utils.GetMyPath(), "certs", "utm.crt")
	VersionPath         = filepath.Join(utils.GetMyPath(), "version.json")
	UpdaterSelfLinux    = "utmstack_updater_self"

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
	DataTypeCiscoGeneric        DataType = "cisco"
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
		DataTypeCiscoGeneric:   {UDP: "514", TCP: "1470"},
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

	ProhibitedPortsChange = []DataType{DataTypeCiscoGeneric, DataTypeNetflow}
)

func ValidateModuleType(typ string) string {
	switch DataType(typ) {
	case DataTypeSyslog, DataTypeVmware, DataTypeEset, DataTypeKaspersky, DataTypeFortinet, DataTypePaloalto,
		DataTypeMikrotik, DataTypeSophosXG, DataTypeSonicwall, DataTypeSentinelOne, DataTypeCiscoGeneric,
		DataTypeDeceptivebytes, DataTypeAix, DataTypePfsense, DataTypeFortiweb, DataTypeSuricata:
		return "syslog"
	case DataTypeNetflow:
		return "netflow"
	case DataTypeWindowsAgent, DataTypeLinuxAgent, DataTypeTraefikModule, DataTypeMongodbModule, DataTypeMysqlModule, DataTypePostgresqlModule,
		DataTypeRedisModule, DataTypeElasticsearchModule, DataTypeKafkaModule, DataTypeKibanaModule, DataTypeLogstashModule, DataTypeNatsModule,
		DataTypeOsqueryModule, DataTypeLinuxAuditdModule, DataTypeHaproxyModule, DataTypeNginxModule, DataTypeIisModule, DataTypeApacheModule:
		return "beats"
	default:
		return "nil"
	}
}
