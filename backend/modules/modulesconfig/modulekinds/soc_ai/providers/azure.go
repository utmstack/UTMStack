package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type AzureProvider struct {
	AbstractProvider
}

func NewAzureProvider(url, model, authType string, customHeaders map[string]string) IProvider {
	return &AzureProvider{*NewAbstractProvider(url, model, authType, customHeaders)}
}

func (p *AzureProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
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
	status, _, err := p.performTestRequest(ctx)
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
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve Azure OpenAI API host. Please verify the API URL is correct.")
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to Azure OpenAI timed out. Please verify the API URL is accessible from this server.")
	}
	return fmt.Errorf("Cannot connect to Azure OpenAI. Please verify the API URL and API Key are correct.")
}
