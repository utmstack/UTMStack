package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type AzureProvider struct {
	AbstractProvider
}

func NewAzureProvider(
	URL string,
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider(URL, Model, AuthType, CustomHeaders)

	return &AzureProvider{
		AbstractProvider: *base,
	}
}

func (p *AzureProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return err
	}
	if p.URL == "" {
		return fmt.Errorf("API URL is required for Azure OpenAI provider")
	}
	if p.Model == "" {
		return fmt.Errorf("Model is required for Azure OpenAI provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Azure OpenAI. Please provide your Azure OpenAI API Key.")
	}

	return p.testConnection()
}

func (p *AzureProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Azure OpenAI API Key. Please verify your api-key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Azure OpenAI API Key does not have the required permissions (HTTP 403). Please verify the API Key has access to the chat completions endpoint.")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Azure OpenAI endpoint not found (HTTP 404). Please verify your Endpoint URL includes the correct resource name and deployment.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("Azure OpenAI API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota. Please check your account billing/usage.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Azure OpenAI API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Azure OpenAI timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Azure OpenAI. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
