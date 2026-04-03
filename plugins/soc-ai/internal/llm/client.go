package llm

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/correlation"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

// LLM retry constants are defined in config package:
// config.LLM_MAX_RETRIES, config.LLM_RETRY_DELAY

// isAnthropicProvider detects if the URL is for Anthropic API
func isAnthropicProvider(url string) bool {
	return strings.Contains(url, "anthropic.com")
}

// SendRequest sends an alert to the configured LLM for analysis
func SendRequest(alert *schema.AlertFields) error {
	content := config.LLM_INSTRUCTION
	if alert == nil {
		return fmt.Errorf("SendRequest: alert is nil")
	}
	correlationContext, err := correlation.GetCorrelationContext(*alert)
	if err != nil {
		return fmt.Errorf("error getting correlation context: %v", err)
	}
	if correlationContext != "" {
		content = fmt.Sprintf("%s%s", content, correlationContext)
	}

	jsonContent, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("error marshalling alert: %v", err)
	}

	// Truncate content if it exceeds the limit to avoid exceeding model context
	jsonStr := string(jsonContent)
	if len(jsonStr) > config.MAX_ALERT_CONTENT_SIZE {
		catcher.Info("Alert content truncated for LLM", map[string]any{
			"process":       "plugin_com.utmstack.soc-ai",
			"alert_id":      alert.Id,
			"original_size": len(jsonStr),
			"truncated_to":  config.MAX_ALERT_CONTENT_SIZE,
		})
		jsonStr = jsonStr[:config.MAX_ALERT_CONTENT_SIZE] + "...[TRUNCATED]"
	}

	cfg := config.GetConfig()
	isAnthropic := isAnthropicProvider(cfg.URL)

	var requestJson []byte

	if isAnthropic {
		// Anthropic uses a different request format
		maxTokens := cfg.MaxTokens
		if maxTokens == 0 {
			maxTokens = 4096 // Default for Anthropic
		}
		req := schema.AnthropicRequest{
			Model:     cfg.Model,
			System:    content,
			MaxTokens: maxTokens,
			Messages: []schema.AnthropicMessage{
				{
					Role:    "user",
					Content: jsonStr,
				},
			},
		}
		requestJson, err = json.Marshal(req)
	} else {
		// OpenAI-compatible format
		req := schema.GPTRequest{
			Model: cfg.Model,
			Messages: []schema.GPTMessage{
				{
					Role:    "system",
					Content: content,
				},
				{
					Role:    "user",
					Content: jsonStr,
				},
			},
		}
		if cfg.MaxTokens > 0 {
			req.MaxTokens = cfg.MaxTokens
		}
		requestJson, err = json.Marshal(req)
	}

	if err != nil {
		return fmt.Errorf("error marshalling request: %v", err)
	}

	headers := buildHeaders(cfg)

	var lastErr error
	backoff := time.Duration(config.LLM_RETRY_DELAY) * time.Second

	for attempt := 1; attempt <= config.LLM_MAX_RETRIES; attempt++ {
		var responseContent string
		var status int

		if isAnthropic {
			// Parse Anthropic response format
			response, s, reqErr := utils.DoParseReq[schema.AnthropicResponse](cfg.URL, requestJson, "POST", headers, config.HTTP_GPT_TIMEOUT)
			status = s
			err = reqErr
			if err == nil && len(response.Content) > 0 {
				responseContent = response.Content[0].Text
			}
		} else {
			// Parse OpenAI-compatible response format
			response, s, reqErr := utils.DoParseReq[schema.GPTResponse](cfg.URL, requestJson, "POST", headers, config.HTTP_GPT_TIMEOUT)
			status = s
			err = reqErr
			if err == nil && len(response.Choices) > 0 {
				responseContent = response.Choices[0].Message.Content
			}
		}

		if err == nil && responseContent != "" {
			err = processResponse(alert, responseContent)
			if err != nil {
				return fmt.Errorf("error processing LLM response: %v", err)
			}
			return nil
		}

		// Handle specific error codes
		switch status {
		case 401:
			return fmt.Errorf("invalid api-key")
		case 429:
			// Rate limited - use exponential backoff
			catcher.Info("LLM rate limited (429), backing off", map[string]any{
				"process":  "plugin_com.utmstack.soc-ai",
				"attempt":  attempt,
				"backoff":  backoff.String(),
				"alert_id": alert.Id,
			})
		case 500, 502, 503, 504:
			// Server errors - worth retrying
			catcher.Info("LLM server error, retrying", map[string]any{
				"process":  "plugin_com.utmstack.soc-ai",
				"status":   status,
				"attempt":  attempt,
				"alert_id": alert.Id,
			})
		}

		lastErr = fmt.Errorf("attempt %d failed: %v (status: %d)", attempt, err, status)

		if attempt < config.LLM_MAX_RETRIES {
			time.Sleep(backoff)
			// Exponential backoff: 2s -> 4s -> 8s (capped at 30s)
			backoff = min(backoff*2, 30*time.Second)
		}
	}

	catcher.Error(fmt.Sprintf("LLM appears to be DOWN - all %d attempts failed for alert %s. URL: %s, Last error: %v",
		config.LLM_MAX_RETRIES, alert.Id, cfg.URL, lastErr), nil, map[string]any{"process": "plugin_com.utmstack.soc-ai"})

	return fmt.Errorf("all attempts to call LLM failed: %v", lastErr)
}

func processResponse(alert *schema.AlertFields, response string) error {
	jsonStr, err := extractJSON(response)
	if err != nil {
		return fmt.Errorf("error extracting json: %v", err)
	}

	alertResponse, err := utils.ConvertFromJsonToStruct[schema.GPTAlertResponse](jsonStr)
	if err != nil {
		return fmt.Errorf("error converting json to struct: %v", err)
	}

	// Validate and normalize the LLM response
	if err := alertResponse.Validate(); err != nil {
		return fmt.Errorf("invalid LLM response: %v", err)
	}

	nextSteps := []string{}
	for _, step := range alertResponse.NextSteps {
		nextSteps = append(nextSteps, fmt.Sprintf("%s:: %s", step.Action, step.Details))
	}

	alert.GPTTimestamp = time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00")
	alert.GPTClassification = alertResponse.Classification
	alert.GPTReasoning = strings.Join(alertResponse.Reasoning, config.LOGS_SEPARATOR)
	alert.GPTNextSteps = strings.Join(nextSteps, "\n")

	return nil
}

// buildHeaders constructs the HTTP headers for the LLM request
func buildHeaders(cfg *config.Config) map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Add Anthropic-specific headers if needed
	if isAnthropicProvider(cfg.URL) {
		headers["anthropic-version"] = config.ANTHROPIC_API_VERSION
	}

	// Add custom headers (includes auth headers configured by frontend)
	if cfg.AuthType == "custom-headers" {
		maps.Copy(headers, cfg.CustomHeaders)
	}
	// If authType is "none", no additional headers are added

	return headers
}

// extractJSON extracts a JSON object from a string that may contain
// surrounding text (common with LLM responses that include explanations)
func extractJSON(s string) (string, error) {
	s = strings.TrimSpace(s)

	// Try to find JSON object boundaries
	start := strings.Index(s, "{")
	if start == -1 {
		return "", fmt.Errorf("no JSON object found in LLM response (missing opening brace)")
	}

	// Find matching closing brace by counting braces
	depth := 0
	end := -1
	for i := start; i < len(s); i++ {
		switch s[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				end = i
				break
			}
		}
		if end != -1 {
			break
		}
	}

	if end == -1 {
		return "", fmt.Errorf("no valid JSON object found in LLM response (unmatched braces)")
	}

	jsonStr := s[start : end+1]

	// Quick validation: try to parse as generic JSON
	var test map[string]any
	if err := json.Unmarshal([]byte(jsonStr), &test); err != nil {
		return "", fmt.Errorf("extracted text is not valid JSON: %v", err)
	}

	return jsonStr, nil
}
