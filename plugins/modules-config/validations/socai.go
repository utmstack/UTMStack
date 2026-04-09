package validations

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"strings"

	"github.com/threatwinds/go-sdk/utils"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

// isAnthropicProvider detects if the URL is for Anthropic API
func isAnthropicProvider(url string) bool {
	return strings.Contains(url, "anthropic.com")
}

// SOCAIConfig holds the parsed SOC-AI configuration
type SOCAIConfig struct {
	AutoAnalyze       bool
	IncidentCreation  bool
	ChangeAlertStatus bool
	Provider          string
	URL               string
	Model             string
	AuthType          string            // "custom-headers", "none"
	MaxTokens         string
	CustomHeaders     map[string]string // All headers including auth (from frontend)
}

var providerDefaultURLs = map[string]string{
	"openai":    "https://api.openai.com/v1/chat/completions",
	"anthropic": "https://api.anthropic.com/v1/messages",
	"azure":     "",
	"gemini":    "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
	"ollama":    "http://localhost:11434/v1/chat/completions",
	"mistral":   "https://api.mistral.ai/v1/chat/completions",
	"deepseek":  "https://api.deepseek.com/chat/completions",
	"groq":      "https://api.groq.com/openai/v1/chat/completions",
}

func ValidateSOCAIConfig(cfg *config.ModuleGroup) error {
	if cfg == nil {
		return fmt.Errorf("SOC_AI configuration is nil")
	}

	socai := parseSOCAIConfig(cfg)

	// Validate required fields
	if socai.URL == "" {
		return fmt.Errorf("URL is required in SOC_AI configuration")
	}
	if socai.Model == "" {
		return fmt.Errorf("Model is required in SOC_AI configuration")
	}

	// Validate authType (optional - defaults to "none" if not specified)
	if socai.AuthType == "" {
		socai.AuthType = "none"
	}
	if socai.AuthType != "custom-headers" && socai.AuthType != "none" {
		return fmt.Errorf("invalid authType '%s', must be 'custom-headers' or 'none'", socai.AuthType)
	}

	// Validate required fields based on authType
	if socai.AuthType == "custom-headers" && len(socai.CustomHeaders) == 0 {
		return fmt.Errorf("Custom Headers are required when authType is 'custom-headers'")
	}

	// Anthropic requires maxTokens
	if isAnthropicProvider(socai.URL) && socai.MaxTokens == "" {
		return fmt.Errorf("Max Tokens is required for Anthropic provider")
	}

	// Test connection
	if err := testSOCAIConnection(socai); err != nil {
		return err
	}

	return nil
}

func parseSOCAIConfig(cfg *config.ModuleGroup) SOCAIConfig {
	socai := SOCAIConfig{
		AuthType:      "none", // default - auth via custom headers or apiKey
		CustomHeaders: make(map[string]string),
	}

	for _, cnf := range cfg.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "utmstack.socai.autoAnalyze":
			socai.AutoAnalyze = cnf.ConfValue == "true"
		case "utmstack.socai.incidentCreation":
			socai.IncidentCreation = cnf.ConfValue == "true"
		case "utmstack.socai.changeAlertStatus":
			socai.ChangeAlertStatus = cnf.ConfValue == "true"
		case "utmstack.socai.provider":
			socai.Provider = cnf.ConfValue
		case "utmstack.socai.url":
			socai.URL = cnf.ConfValue
		case "utmstack.socai.model":
			socai.Model = cnf.ConfValue
		case "utmstack.socai.authType":
			if cnf.ConfValue != "" {
				socai.AuthType = cnf.ConfValue
			}
		case "utmstack.socai.maxTokens":
			socai.MaxTokens = cnf.ConfValue
		case "utmstack.socai.customHeaders":
			if cnf.ConfValue != "" {
				if err := json.Unmarshal([]byte(cnf.ConfValue), &socai.CustomHeaders); err != nil {
					fmt.Printf("Warning: Failed to parse customHeaders JSON: %v\n", err)
				}
			}
		}
	}

	// Resolve URL from provider if not custom
	if socai.Provider != "" && socai.Provider != "custom" {
		if defaultURL, ok := providerDefaultURLs[socai.Provider]; ok && defaultURL != "" {
			socai.URL = defaultURL
		}
	}

	return socai
}

func testSOCAIConnection(socai SOCAIConfig) error {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Add custom headers (includes auth headers configured by frontend)
	if socai.AuthType == "custom-headers" {
		maps.Copy(headers, socai.CustomHeaders)
	}
	// If authType is "none", no additional headers are added

	// Test connection with GET request (most APIs return error but validate auth)
	response, status, err := utils.DoReq[map[string]any](socai.URL, nil, "GET", headers, false)

	// Handle response
	switch status {
	case http.StatusOK, http.StatusMethodNotAllowed, http.StatusBadRequest:
		// These are acceptable - means we reached the API and auth worked
		// 405 = endpoint doesn't accept GET but auth passed
		// 400 = bad request but auth passed
		return nil
	case http.StatusUnauthorized:
		return fmt.Errorf("SOC_AI API Key is invalid (401 Unauthorized)")
	case http.StatusForbidden:
		return fmt.Errorf("SOC_AI API Key does not have permission (403 Forbidden)")
	case http.StatusRequestTimeout:
		return fmt.Errorf("SOC_AI connection timed out")
	case http.StatusNotFound:
		return fmt.Errorf("SOC_AI URL not found (404) - check the URL is correct")
	default:
		if err != nil {
			return fmt.Errorf("SOC_AI connection failed: %v", err)
		}
		fmt.Printf("SOC_AI validation: status %d, response: %v\n", status, response)
		return nil // Accept other status codes as potentially valid
	}
}
