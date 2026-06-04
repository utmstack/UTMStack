package dto

type ResolveFilterValuesResponse struct {
	AgentPlatforms []string `json:"agentPlatform"`
	Users          []string `json:"users"`
}
