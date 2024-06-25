package config

import (
	"os"
	"time"
)

func KeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/DeleteAgent",

		"/agent.CollectorService/CollectorStream",
		"/agent.CollectorService/DeleteCollector",
		"/agent.CollectorService/GetCollectorConfig",

		"/agent.PingService/Ping",

		"/agent.UpdatesService/CheckAgentUpdates",
		"/agent.UpdatesService/CheckCollectorUpdates",
	}
}

func ConnectionKeyRoutes() []string {
	return []string{
		"/agent.AgentService/RegisterAgent",
		"/agent.CollectorService/RegisterCollector",

		"/agent.UpdatesService/CheckAgentUpdates",
		"/agent.UpdatesService/CheckCollectorUpdates",
	}
}

func InternalKeyRoutes() []string {
	return []string{
		"/agent.AgentService/ListAgents",
		"/agent.AgentService/UpdateAgentType",  // Not used
		"/agent.AgentService/UpdateAgentGroup", // Not used
		"/agent.AgentService/ListAgentCommands",
		"/agent.AgentService/GetAgentByHostname",
		"/agent.AgentService/ListAgentsWithCommands",

		"/agent.CollectorService/ListCollector",
		"/agent.CollectorService/ListCollectorHostnames",
		"/agent.CollectorService/GetCollectorsByHostnameAndModule",

		"/agent.PanelService/ProcessCommand",
		"/agent.PanelCollectorService/RegisterCollectorConfig",
	}
}

const (
	PanelConnectionKeyUrl = "%s/api/authenticateFederationServiceManager"
	UTMSharedKeyEnv       = "INTERNAL_KEY"
	UTMEncryptionKeyEnv   = "ENCRYPTION_KEY"
	UTMPanelServiceEnv    = "PANEL_SERV_NAME"
	UTMHostEnv            = "UTM_HOST"
	ProxyAPIKeyHeader     = "Utm-Connection-Key"
	MASTERVERSIONENDPOINT = "/management/info"
	Bucket                = "https://cdn.utmstack.com/"
	VOLPATH               = "/data"
	VersionsFile          = "versions.json"
	CHECK_EVERY           = 5 * time.Minute
)

func GetInternalKey() string {
	return os.Getenv(UTMSharedKeyEnv)
}

func GetPanelServiceName() string {
	return os.Getenv(UTMPanelServiceEnv)
}
