package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type CustomProvider struct {
	AbstractProvider
}

func NewCustomProvider(url, model, authType string, customHeaders map[string]string) IProvider {
	return &CustomProvider{*NewAbstractProvider(url, model, authType, customHeaders)}
}

func (p *CustomProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
		return err
	}
	if p.URL == "" {
		return fmt.Errorf("API URL is required for Custom provider")
	}
	if p.Model == "" {
		return fmt.Errorf("Model is required for Custom provider")
	}
	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("Custom Headers are required for Custom. Please provide your Custom API Key.")
	}
	status, _, err := p.performTestRequest(ctx)
	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid API Key for Custom provider. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Custom API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Custom API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Custom API rate limit exceeded (HTTP 429).")
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve Custom API host. Please verify the API URL is correct.")
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to Custom provider timed out. Please verify the API URL is accessible from this server.")
	}
	return fmt.Errorf("Cannot connect to Custom provider. Please verify the API URL is correct.")
}
