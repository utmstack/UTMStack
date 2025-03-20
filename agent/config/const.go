package config

import (
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/utils"
)

const REPLACE_KEY string = ""

type DataType string

type ProtoPort struct {
	UDP string
	TCP string
}

var (
	DependUrl        = "https://%s:%s/private/dependencies/agent/%s"
	AgentManagerPort = "9000"
	LogAuthProxyPort = "50051"
	DependenciesPort = "9001"

	ServiceLogFile       = filepath.Join(utils.GetMyPath(), "logs", "utmstack_agent.log")
	ModulesServName      = "UTMStackModulesLogsCollector"
	WinServName          = "UTMStackWindowsLogsCollector"
	CollectorFileName    = filepath.Join(utils.GetMyPath(), "log-collector-config.json")
	CollectorFileNameOld = filepath.Join(utils.GetMyPath(), "log-collector-configuration.json")
	UUIDFileName         = filepath.Join(utils.GetMyPath(), "uuid.yml")
	ConfigurationFile    = filepath.Join(utils.GetMyPath(), "config.yml")
	PortRangeMin         = "7000"
	PortRangeMax         = "9000"
	RetentionConfigFile  = filepath.Join(utils.GetMyPath(), "retention.json")
	LogsDBFile           = filepath.Join(utils.GetMyPath(), "logs_process", "logs.db")
	VersionPath          = filepath.Join(utils.GetMyPath(), "version.json")
	MESSAGE_HEADER       = "utm_stack_agent_ds"
	BatchCapacity        = 100

	// MaxConnectionTime     = 120 * time.Second
	// SERV_NAME            = "UTMStackAgent"
	// SERV_LOG             = "utmstack_agent.log"
	// ModulesLockName      = "utmstack_modules_collector.lock"
	// WinLockName          = "utmstack_windows_collector.lock"
	// RedlineLockName      = "utmstack_redline.lock"
	// RedlineServName      = "UTMStackRedline"
	// BatchToSend          = 5

	DataTypeWindowsAgent        DataType = "beats_windows_agent"
	DataTypeSyslog              DataType = "syslog"
	DataTypeVmware              DataType = "vmware"
	DataTypeLinuxAgent          DataType = "beats_linux_agent"
	DataTypeEset                DataType = "antivirus_eset"
	DataTypeKaspersky           DataType = "antivirus_kaspersky"
	DataTypeTraefikModule       DataType = "beats_traefik_module"
	DataTypeMongodbModule       DataType = "beats_mongodb_module"
	DataTypeMysqlModule         DataType = "beats_mysql_module"
	DataTypePostgresqlModule    DataType = "beats_postgresql_module"
	DataTypeRedisModule         DataType = "beats_redis_module"
	DataTypeElasticsearchModule DataType = "beats_elasticsearch_module"
	DataTypeKafkaModule         DataType = "beats_kafka_module"
	DataTypeKibanaModule        DataType = "beats_kibana_module"
	DataTypeLogstashModule      DataType = "beats_logstash_module"
	DataTypeCiscoAsa            DataType = "cisco_asa"
	DataTypeCiscoMeraki         DataType = "cisco_meraki"
	DataTypeFortinet            DataType = "firewall_fortinet"
	DataTypePaloalto            DataType = "firewall_paloalto"
	DataTypeMikrotik            DataType = "firewall_mikrotik"
	DataTypeCiscoFirepower      DataType = "cisco_firepower"
	DataTypeSophosXG            DataType = "firewall_sophos"
	DataTypeCiscoSwitch         DataType = "cisco_switch"
	DataTypeSonicwall           DataType = "firewall_sonicwall"
	DataTypeNatsModule          DataType = "beats_nats_module"
	DataTypeDeceptivebytes      DataType = "antivirus_deceptivebytes"
	DataTypeOsqueryModule       DataType = "beats_osquery_module"
	DataTypeLinuxAuditdModule   DataType = "beats_auditd_module"
	DataTypeHaproxyModule       DataType = "beats_haproxy_module"
	DataTypeNginxModule         DataType = "beats_nginx_module"
	DataTypeIisModule           DataType = "beats_iis_module"
	DataTypeApacheModule        DataType = "beats_apache_module"
	DataTypeSentinelOne         DataType = "antivirus_sentinel_one"
	DataTypeCiscoGeneric        DataType = "cisco"
	DataTypeMacOs               DataType = "macos_logs"
	DataTypeGeneric             DataType = "generic"
	DataTypeNetflow             DataType = "netflow"
	DataTypeAix                 DataType = "ibm_aix"
	DataTypePfsense             DataType = "firewall_pfsense"
	DataTypeFortiweb            DataType = "firewall_fortiweb"

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
		DataTypeMacOs:          {UDP: "7015", TCP: "7015"},
		DataTypeAix:            {UDP: "7016", TCP: "7016"},
		DataTypePfsense:        {UDP: "7017", TCP: "7017"},
		DataTypeFortiweb:       {UDP: "7018", TCP: "7018"},
		DataTypeNetflow:        {UDP: "2055", TCP: ""},
	}

	ProhibitedPortsChange = []DataType{DataTypeCiscoGeneric, DataTypeNetflow}
)

func GetMessageFormated(host string, msg string) string {
	return "[" + MESSAGE_HEADER + "=" + host + "]-" + msg
}

func ValidateModuleType(typ string) string {
	switch DataType(typ) {
	case DataTypeSyslog, DataTypeVmware, DataTypeEset, DataTypeKaspersky, DataTypeFortinet, DataTypePaloalto,
		DataTypeMikrotik, DataTypeSophosXG, DataTypeSonicwall, DataTypeSentinelOne, DataTypeCiscoGeneric, DataTypeMacOs,
		DataTypeDeceptivebytes, DataTypeAix, DataTypePfsense, DataTypeFortiweb:
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
