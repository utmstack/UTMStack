package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type OllamaProvider struct {
	AbstractProvider
}

func NewOllamaProvider(url, model, authType string, customHeaders map[string]string) IProvider {
	return &OllamaProvider{*NewAbstractProvider(url, model, authType, customHeaders)}
}

func (p *OllamaProvider) Validate(ctx context.Context) error {
	if err := p.validateAuthType(); err != nil {
		return fmt.Errorf("Ollama validation error: %v", err)
	}
	if p.URL == "" {
		return fmt.Errorf("Ollama validation error: Ollama server URL could not be determined. Please verify the provider configuration.")
	}
	status, _, err := p.performTestRequest(ctx)
	switch status {
	case http.StatusOK, http.StatusBadRequest:
		return nil
	case http.StatusNotFound, http.StatusMethodNotAllowed:
		return fmt.Errorf("Ollama API not found at '%s' (HTTP 404). Please verify Ollama is running and the URL is correct.", p.URL)
	case http.StatusRequestTimeout:
		return fmt.Errorf("Connection to Ollama timed out. Please verify Ollama is running at '%s' and is accessible from this server.", p.URL)
	}
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "no such host"), strings.Contains(msg, "lookup"):
		return fmt.Errorf("Cannot resolve Ollama server at '%s'. Please verify the hostname is correct and accessible.", p.URL)
	case strings.Contains(msg, "connection refused"):
		return fmt.Errorf("Connection refused by Ollama at '%s'. Please verify Ollama is running.", p.URL)
	case strings.Contains(msg, "timeout"), strings.Contains(msg, "deadline"):
		return fmt.Errorf("Connection to Ollama timed out. Please verify Ollama is running at '%s' and is accessible from this server.", p.URL)
	}
	return fmt.Errorf("Cannot connect to Ollama at '%s'. Please verify Ollama is running.", p.URL)
}
