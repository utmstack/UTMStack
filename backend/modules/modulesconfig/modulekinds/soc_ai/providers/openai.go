package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type OpenAIProvider struct {
	AbstractProvider
}

func NewOpenAIProvider(model, authType string, customHeaders map[string]string) IProvider {
	return &OpenAIProvider{*NewAbstractProvider("https://api.openai.com/v1/chat/completions", model, authType, customHeaders)}
}

func (p *OpenAIProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
		return fmt.Errorf("OpenAI validation error: %v", err)
	}
	if p.Model == "" {
		return fmt.Errorf("Model is required for OpenAI provider")
	}
	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for OpenAI. Please provide your OpenAI API Key.")
	}
	status, _, err := p.performTestRequest(ctx)
	return mapStatus(err, status, "OpenAI", "OpenAI API")
}

func mapStatus(err error, status int, vendor, vendorAPI string) error {
	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid %s API Key. Please verify your API Key is correct.", vendor)
	case http.StatusForbidden:
		return fmt.Errorf("%s API Key does not have the required permissions (HTTP 403). Please verify the API Key has access to the chat completions endpoint.", vendor)
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("%s endpoint not found (HTTP 404). Please verify the API URL is correct.", vendorAPI)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%s rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.", vendorAPI)
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve %s host. Please verify the API URL is correct.", vendorAPI)
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to %s timed out. Please verify the API URL is accessible from this server.", vendor)
	}
	return fmt.Errorf("Cannot connect to %s. Please verify the API URL and API Key are correct.", vendor)
}
