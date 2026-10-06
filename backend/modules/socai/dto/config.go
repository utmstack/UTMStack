package dto

const MaskedValue = "*****"

type ConfigRequest struct {
	Provider          string            `json:"provider" binding:"required"`
	Model             string            `json:"model" binding:"required"`
	URL               string            `json:"url"`
	APIKey            string            `json:"apiKey"`
	AuthType          string            `json:"authType"`
	AuthHeaderName    string            `json:"authHeaderName"`
	CustomHeaders     map[string]string `json:"customHeaders"`
	MaxTokens         int               `json:"maxTokens"`
	MaxToolIterations int               `json:"maxToolIterations"`
	// AutoAnalyze: triage every new alert automatically.
	AutoAnalyze bool `json:"autoAnalyze"`
	// AllowIncidents: the triage agent may create/update incidents when it
	// decides an alert warrants one. Everything else the agent can do
	// (dashboards, compliance, correlation, datasources, SOAR) only matters
	// to the interactive chat, where a person is already driving each
	// conversation, so it needs no separate switch — configuring SOC-AI at
	// all is the signal that it's meant to be used.
	AllowIncidents bool `json:"allowIncidents"`
}

type ConfigResponse struct {
	Configured bool `json:"configured"`

	// Inherited reports that this is the instance default rather than the
	// tenant's own, so the UI can say so and offer to stop inheriting — or, if
	// it already has its own, offer to go back.
	Inherited bool `json:"inherited"`

	Provider          string            `json:"provider"`
	Model             string            `json:"model"`
	URL               string            `json:"url"`
	APIKeySet         bool              `json:"apiKeySet"`
	AuthType          string            `json:"authType"`
	AuthHeaderName    string            `json:"authHeaderName"`
	CustomHeaders     map[string]string `json:"customHeaders"`
	MaxTokens         int               `json:"maxTokens"`
	MaxToolIterations int               `json:"maxToolIterations"`
	AutoAnalyze       bool              `json:"autoAnalyze"`
	AllowIncidents    bool              `json:"allowIncidents"`
}
