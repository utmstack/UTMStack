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

var providerDisplayNames = map[string]string{
	"openai":    "OpenAI",
	"anthropic": "Anthropic",
	"azure":     "Azure OpenAI",
	"gemini":    "Google Gemini",
	"ollama":    "Ollama",
	"mistral":   "Mistral AI",
	"deepseek":  "DeepSeek",
	"groq":      "Groq",
	"custom":    "Custom",
}

func getProviderName(provider string) string {
	if name, ok := providerDisplayNames[provider]; ok {
		return name
	}
	return provider
}

func ValidateSOCAIConfig(cfg *config.ModuleGroup) error {
	if cfg == nil {
		return fmt.Errorf("SOC AI configuration is not provided")
	}

	socai := parseSOCAIConfig(cfg)
	providerName := getProviderName(socai.Provider)

	// Validate required fields
	if socai.URL == "" {
		if socai.Provider == "custom" || socai.Provider == "azure" || socai.Provider == "ollama" {
			return fmt.Errorf("API URL is required for %s provider", providerName)
		}
		return fmt.Errorf("API URL could not be determined. Please verify the provider configuration.")
	}
	if socai.Model == "" {
		return fmt.Errorf("Model is required for %s provider", providerName)
	}

	// Validate authType
	if socai.AuthType == "" {
		socai.AuthType = "none"
	}
	if socai.AuthType != "custom-headers" && socai.AuthType != "none" {
		return fmt.Errorf("Invalid authentication type '%s'. Must be 'custom-headers' or 'none'.", socai.AuthType)
	}

	// Validate auth headers exist when needed
	if socai.AuthType == "custom-headers" && len(socai.CustomHeaders) == 0 {
		if socai.Provider == "ollama" {
			// Ollama doesn't need auth
		} else {
			return fmt.Errorf("API Key is required for %s. Please provide your %s API Key.", providerName, providerName)
		}
	}

	// Anthropic requires maxTokens
	if isAnthropicProvider(socai.URL) && socai.MaxTokens == "" {
		return fmt.Errorf("Max Tokens is required for Anthropic. Please set a value (e.g., 4096).")
	}

	// Test connection
	if err := testSOCAIConnection(socai); err != nil {
		return err
	}

	return nil
}

func parseSOCAIConfig(cfg *config.ModuleGroup) SOCAIConfig {
	socai := SOCAIConfig{
		AuthType:      "none",
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
	providerName := getProviderName(socai.Provider)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Add custom headers (includes auth headers configured by frontend)
	if socai.AuthType == "custom-headers" {
		maps.Copy(headers, socai.CustomHeaders)
	}

	// Test connection with GET request (most APIs return error but validate auth)
	_, status, err := utils.DoReq[map[string]any](socai.URL, nil, "GET", headers, false)

	switch status {
	case http.StatusOK, http.StatusMethodNotAllowed, http.StatusBadRequest:
		// These are acceptable - means we reached the API and auth worked
		return nil
	case http.StatusUnauthorized:
		if socai.Provider == "anthropic" {
			return fmt.Errorf("Invalid Anthropic API Key. Please verify your x-api-key is correct.")
		}
		if socai.Provider == "azure" {
			return fmt.Errorf("Invalid Azure OpenAI API Key. Please verify your api-key is correct.")
		}
		return fmt.Errorf("Invalid API Key for %s. Please verify your API Key is correct.", providerName)
	case http.StatusForbidden:
		return fmt.Errorf("%s API Key does not have the required permissions (HTTP 403). Please verify the API Key has access to the chat completions endpoint.", providerName)
	case http.StatusNotFound:
		if socai.Provider == "azure" {
			return fmt.Errorf("Azure OpenAI endpoint not found (HTTP 404). Please verify your Endpoint URL includes the correct resource name and deployment.")
		}
		if socai.Provider == "ollama" {
			return fmt.Errorf("Ollama API not found at '%s' (HTTP 404). Please verify Ollama is running and the URL is correct.", socai.URL)
		}
		return fmt.Errorf("%s API endpoint not found (HTTP 404). Please verify the API URL is correct.", providerName)
	case http.StatusRequestTimeout:
		if socai.Provider == "ollama" {
			return fmt.Errorf("Connection to Ollama timed out. Please verify Ollama is running at '%s' and is accessible from this server.", socai.URL)
		}
		return fmt.Errorf("Connection to %s timed out. Please verify the API URL is accessible from this server.", providerName)
	case http.StatusTooManyRequests:
		return fmt.Errorf("%s API rate limit exceeded (HTTP 429). Your API Key may have exceeded its quota. Please check your %s account billing/usage.", providerName, providerName)
	default:
		if err != nil {
			errMsg := strings.ToLower(err.Error())
			if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
				if socai.Provider == "ollama" {
					return fmt.Errorf("Cannot resolve Ollama server at '%s'. Please verify the hostname is correct and accessible.", socai.URL)
				}
				return fmt.Errorf("Cannot resolve %s API host. Please verify the API URL is correct.", providerName)
			}
			if strings.Contains(errMsg, "connection refused") {
				if socai.Provider == "ollama" {
					return fmt.Errorf("Connection refused by Ollama at '%s'. Please verify Ollama is running.", socai.URL)
				}
				return fmt.Errorf("Connection refused by %s. Please verify the API URL is correct.", providerName)
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
				return fmt.Errorf("Connection to %s timed out. Please verify the API URL is accessible from this server.", providerName)
			}
			return fmt.Errorf("Cannot connect to %s. Please verify the API URL and API Key are correct.", providerName)
		}
		return nil // Accept other status codes as potentially valid
	}
}
