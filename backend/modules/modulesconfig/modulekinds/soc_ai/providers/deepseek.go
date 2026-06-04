package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type DeepSeekProvider struct {
	AbstractProvider
}

func NewDeepSeekProvider(model, authType string, customHeaders map[string]string) IProvider {
	return &DeepSeekProvider{*NewAbstractProvider("https://api.deepseek.com/chat/completions", model, authType, customHeaders)}
}

func (p *DeepSeekProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
		return err
	}
	if p.Model == "" {
		return fmt.Errorf("Model is required for DeepSeek provider")
	}
	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for DeepSeek. Please provide your DeepSeek API Key.")
	}
	status, _, err := p.performTestRequest(ctx)
	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid DeepSeek API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("DeepSeek API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("DeepSeek API endpoint not found (HTTP 404). Please verify the API URL is correct.")
	case http.StatusTooManyRequests:
		return fmt.Errorf("DeepSeek API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.")
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve DeepSeek API host. Please verify the API URL is correct.")
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to DeepSeek timed out. Please verify the API URL is accessible from this server.")
	}
	return fmt.Errorf("Cannot connect to DeepSeek. Please verify the API URL and API Key are correct.")
}
