package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type GroqProvider struct {
	AbstractProvider
}

func NewGroqProvider(
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider( "https://api.groq.com/openai/v1/chat/completions", Model, AuthType, CustomHeaders)

	return &GroqProvider{
		AbstractProvider: *base,
	}
}

func (p *GroqProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return err
	}

	if p.Model == "" {
		return fmt.Errorf("Model is required for Groq provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Groq. Please provide your Groq API Key.")
	}

	return p.testConnection()
}

func (p *GroqProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Groq API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Groq API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Groq API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Groq API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Groq API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Groq timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Groq. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
