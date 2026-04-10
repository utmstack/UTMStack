package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type CustomProvider struct {
	AbstractProvider
}

func NewCustomProvider(
	URL string,
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider(URL, Model, AuthType, CustomHeaders)

	return &CustomProvider{
		AbstractProvider: *base,
	}
}

func (p *CustomProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
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

	return p.testConnection()
}

func (p *CustomProvider) testConnection() error {
	status, err := p.PerformTestRequest()

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
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Custom API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Custom provider timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Custom provider. Please verify the API URL is correct.")
		}
		return nil
	}
}
