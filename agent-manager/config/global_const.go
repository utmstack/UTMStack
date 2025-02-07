package config

import "os"

func KeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/UpdateAgent",
		"/agent.AgentService/DeleteAgent",

		"/agent.CollectorService/CollectorStream",
		"/agent.CollectorService/DeleteCollector",
		"/agent.CollectorService/GetCollectorConfig",

		"/agent.PingService/Ping",
	}
}

func ConnectionKeyRoutes() []string {
	return []string{
		"/agent.AgentService/RegisterAgent",
		"/agent.CollectorService/RegisterCollector",
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

const PanelConnectionKeyUrl = "%s/api/authenticateFederationServiceManager"
const UTMSharedKeyEnv = "INTERNAL_KEY"
const UTMEncryptionKeyEnv = "ENCRYPTION_KEY"
const UTMHostEnv = "UTM_HOST"

func GetInternalKey() string {
	return os.Getenv(UTMSharedKeyEnv)
}
