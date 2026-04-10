package providers

import (
	"fmt"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	AbstractProvider
}

func NewOllamaProvider(
	URL string,
	Model string,
	AuthType string,
	CustomHeaders map[string]string,
) IProvider {
	base := NewAbstractProvider(URL, Model, AuthType, CustomHeaders)

	return &OllamaProvider{
		AbstractProvider: *base,
	}
}

func (p *OllamaProvider) Validate() error {
	if err := p.AbstractProvider.Validate(); err != nil {
		return fmt.Errorf("Ollama validation error: %v", err)
	}

	if p.URL == "" {
		return fmt.Errorf("Ollama validation error: Ollama server URL could not be determined. Please verify the provider configuration.")
	}
	return p.testConnection()
}

func (p *OllamaProvider) testConnection() error {
	status, err := p.PerformTestRequest()

	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Ollama API not found at '%s' (HTTP 404). Please verify Ollama is running and the URL is correct.", p.URL)
	case http.StatusRequestTimeout:
		return fmt.Errorf("Connection to Ollama timed out. Please verify Ollama is running at '%s' and is accessible from this server.", p.URL)
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				return fmt.Errorf("Cannot resolve Ollama server at '%s'. Please verify the hostname is correct and accessible.", p.URL)
			}
			if strings.Contains(errMsg, "connection refused") {
				return fmt.Errorf("Connection refused by Ollama at '%s'. Please verify Ollama is running.", p.URL)
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to Ollama timed out. Please verify Ollama is running at '%s' and is accessible from this server.", p.URL)
			}
			return fmt.Errorf("Cannot connect to Ollama at '%s'. Please verify Ollama is running.", p.URL)
		}
		return nil
	}
}
