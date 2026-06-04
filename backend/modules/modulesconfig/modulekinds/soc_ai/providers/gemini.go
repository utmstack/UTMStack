package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type GeminiProvider struct {
	AbstractProvider
}

func NewGeminiProvider(model, authType string, customHeaders map[string]string) IProvider {
	return &GeminiProvider{*NewAbstractProvider("https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", model, authType, customHeaders)}
}

func (p *GeminiProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
		return err
	}
	if p.Model == "" {
		return fmt.Errorf("Model is required for Google Gemini provider")
	}
	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Google Gemini. Please provide your Google Gemini API Key.")
	}
	status, body, err := p.performTestRequest(ctx)
	switch status {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		if strings.HasPrefix(body, "API key not valid.") {
			return fmt.Errorf("Invalid Gemini API Key. Please verify your x-api-key is correct.")
		}
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Google Gemini API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Google Gemini API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		if strings.HasSuffix(body, "available models and their supported methods.") {
			return fmt.Errorf("Invalid Gemini model %s.", p.Model)
		}
		return nil
	case http.StatusTooManyRequests:
		return fmt.Errorf("Google Gemini API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.")
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve Google Gemini API host. Please verify the API URL is correct.")
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to Google Gemini timed out. Please verify the API URL is accessible from this server.")
	}
	return fmt.Errorf("Cannot connect to Google Gemini. Please verify the API URL and API Key are correct.")
}
