package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type AnthropicProvider struct {
	AbstractProvider
	MaxTokens string
}

func NewAnthropicProvider(
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
	MaxTokens string,
) IProvider {
	base := NewAbstractProvider("https://api.anthropic.com/v1/messages", Model, AuthType, CustomHeaders)

	return &AnthropicProvider{
		AbstractProvider: *base,
		MaxTokens:        MaxTokens,
	}
}

func (p *AnthropicProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return err
	}

	if p.Model == "" {
		return fmt.Errorf("Model is required for Anthropic provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Anthropic. Please provide your Anthropic API Key.")
	}

	if p.MaxTokens == "" {
		return fmt.Errorf("Max Tokens is required for Anthropic. Please set a value (e.g., 4096).")
	}

	return p.testConnection()
}

func (p *AnthropicProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		// These are acceptable - means we reached the API and auth worked
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Anthropic API Key. Please verify your x-api-key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Anthropic API Key does not have the required permissions (HTTP 403). Please verify the API Key has access to the chat completions endpoint.")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Anthropic API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusRequestTimeout:
		return fmt.Errorf("Connection to Anthropic timed out. Please verify the API URL is accessible from this server.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Anthropic API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota. Please check your Anthropic account billing/usage.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Anthropic API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "connection refused") {
				return fmt.Errorf("Connection refused by Anthropic. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Anthropic timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Anthropic. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
