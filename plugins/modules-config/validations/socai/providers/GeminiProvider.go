package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type GeminiProvider struct {
	AbstractProvider
}

func NewGeminiProvider(
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider( "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions", Model, AuthType, CustomHeaders)

	return &GeminiProvider{
		AbstractProvider: *base,
	}
}

func (p *GeminiProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return err
	}

	if p.Model == "" {
		return fmt.Errorf("Model is required for Google Gemini provider")
	}

	if p.AuthType == "custom-headers" && len(p.CustomHeaders) == 0 {
		return fmt.Errorf("API Key is required for Google Gemini. Please provide your Google Gemini API Key.")
	}

	return p.testConnection()
}

func (p *GeminiProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		if strings.HasPrefix(err.Error(),"API key not valid."){
			return fmt.Errorf("Invalid Gemini API Key. Please verify your x-api-key is correct.")
		}else{
			return nil
		}
	case http.StatusUnauthorized:
		return fmt.Errorf("Invalid Google Gemini API Key. Please verify your API Key is correct.")
	case http.StatusForbidden:
		return fmt.Errorf("Google Gemini API Key does not have the required permissions (HTTP 403).")
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		if strings.HasSuffix(err.Error(),"available models and their supported methods."){
			return fmt.Errorf("Invalid Gemini model %s.",p.Model)
		}else{
			return nil
		}
	case http.StatusTooManyRequests:
		return fmt.Errorf("Google Gemini API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota.")
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Google Gemini API host. Please verify the API URL is correct.")
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Google Gemini timed out. Please verify the API URL is accessible from this server.")
			}
			return fmt.Errorf("Cannot connect to Google Gemini. Please verify the API URL and API Key are correct.")
		}
		return nil
	}
}
