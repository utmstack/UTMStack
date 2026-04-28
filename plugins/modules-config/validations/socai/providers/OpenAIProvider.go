package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type OpenAIProvider struct {
	AbstractProvider
}

func NewOpenAIProvider(
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider("https://api.openai.com/v1/chat/completions", Model, AuthType, CustomHeaders)

	return &OpenAIProvider{
		AbstractProvider: *base,
	}
}

func (p *OpenAIProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return fmt.Errorf("OpenAI validation error: %v", err)
	}

	if p.AbstractProvider.Model == "" {
		return fmt.Errorf("Model is required for OpenAI provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for OpenAI. Please provide your Groq API Key.")
	}
	return p.testConnection()
}

func (p *OpenAIProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid OpenAI API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("OpenAI API Key does not have the required permissions (HTTP 403). Please verify the API Key has access to the chat completions endpoint.")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("OpenAI API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("OpenAI API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota. Please check your OpenAI account billing/usage.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve OpenAI API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to OpenAI timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to OpenAI. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
