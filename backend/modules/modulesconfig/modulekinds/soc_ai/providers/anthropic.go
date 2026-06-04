package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type AnthropicProvider struct {
	AbstractProvider
	MaxTokens string
}

func NewAnthropicProvider(model, authType string, customHeaders map[string]string, maxTokens string) IProvider {
	return &AnthropicProvider{
		AbstractProvider: *NewAbstractProvider("https://api.anthropic.com/v1/messages", model, authType, customHeaders),
		MaxTokens:        maxTokens,
	}
}

func (p *AnthropicProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
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
	status, _, err := p.performTestRequest(ctx)
	switch status {
	case http.StatusOK, http.StatusBadRequest:
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
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve Anthropic API host. Please verify the API URL is correct.")
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("Connection refused by Anthropic. Please verify the API URL is correct.")
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to Anthropic timed out. Please verify the API URL is accessible from this server.")
	}
	return fmt.Errorf("Cannot connect to Anthropic. Please verify the API URL and API Key are correct.")
}
