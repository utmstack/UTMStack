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
	AutoAnalyze       bool              `json:"autoAnalyze"`
	Capabilities      []string          `json:"capabilities"`
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
	Capabilities      []string          `json:"capabilities"`
}
