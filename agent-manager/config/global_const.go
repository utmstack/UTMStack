package config

import "os"

func AgentKeyAuthRoutes() []string {
	return []string{
		"/agent.AgentService/AgentStream",
		"/agent.AgentService/DeleteAgent",
		"/agent.CollectorService/CollectorStream",
		"/agent.CollectorService/DeleteCollector",
	}
}

func ConnectionKeyRoutes() []string {
	return []string{
		"/agent.AgentService/RegisterAgent",
		"/agent.AgentService/ListAgents",
		"/agent.AgentService/ListAgentCommands",
		"/agent.AgentService/GetAgentByHostname",
		"/agent.AgentService/ListAgentsWithCommands",
		"/agent.AgentService/UpdateAgentType",
		"/agent.AgentService/UpdateAgentGroup",
		"/agent.AgentService/ListAgentCommands",

		"/agent.CollectorService/RegisterCollector",
		"/agent.CollectorService/ListCollector",
		"/agent.CollectorService/GetCollectorsByHostnameAndModule",
		"/agent.CollectorService/ListCollectorHostnames",
	}
}

const PanelConnectionKeyUrl = "%s/api/authenticateFederationServiceManager"
const UTMSharedKeyEnv = "INTERNAL_KEY"
const UTMEncryptionKeyEnv = "ENCRYPTION_KEY"
const UTMHostEnv = "UTM_HOST"

func GetInternalKey() string {
	return os.Getenv(UTMSharedKeyEnv)
}
