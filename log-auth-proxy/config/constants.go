package config

import "os"

const PanelConnectionKeyEndpoint = "%s/api/federation-service/token"
const PanelPipelinesEndpoint = "%s/api/logstash-pipelines/port"
const LogstashPipelinesEndpoint = "http://%s:%s"
const PanelAPIKeyHeader = "Utm-Internal-Key"
const ProxyAPIKeyHeader = "Utm-Connection-Key"
const ProxyLogTypeHeader = "Utm-Log-Type"
const ProxySourceHeader = "Utm-Log-Source"
const UTMSharedKeyEnv = "INTERNAL_KEY"
const UTMHostEnv = "UTM_HOST"
const UTMAgentManagerHostEnv = "UTM_AGENT_MANAGER_HOST"
const UTMLogstashHostEnv = "UTM_LOGSTASH_HOST"
const UTMCertsLocationEnv = "UTM_CERTS_LOCATION"
const UTMCertFileName = "utm.crt"
const UTMCertFileKey = "utm.key"
const UTMLogSeparator = "<utm-log-separator>"

func LogstashHost() string {
	return os.Getenv(UTMLogstashHostEnv)
}
