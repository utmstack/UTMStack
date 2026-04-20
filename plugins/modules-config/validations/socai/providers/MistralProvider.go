package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type MistralProvider struct {
	AbstractProvider
}

func NewMistralProvider(
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider( "https://api.mistral.ai/v1/chat/completions", Model, AuthType, CustomHeaders)

	return &MistralProvider{
		AbstractProvider: *base,
	}
}

func (p *MistralProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return fmt.Errorf("Mistral AI validation error: %v", err)
	}

	if p.Model == "" {
		return fmt.Errorf("Model is required for Mistral provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Mistral. Please provide your Groq API Key.")
	}
	return p.testConnection()
}

func (p *MistralProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Mistral AI API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Mistral AI API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Mistral AI API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Mistral AI API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Mistral AI API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Mistral AI timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Mistral AI. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
