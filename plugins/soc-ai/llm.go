package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/correlation"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

func sendRequestToLLM(alert *schema.AlertFields) error {
	const maxRetries = 3
	const retryDelay = 2 * time.Second

	content := config.LLM_INSTRUCTION
	if alert == nil {
		return fmt.Errorf("sendRequestToOpenAI: alert is nil")
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

	req := schema.GPTRequest{
		Model: config.GetConfig().Model,
		Messages: []schema.GPTMessage{
			{
				Role:    "system",
				Content: content,
			},
			{
				Role:    "user",
				Content: string(jsonContent),
			},
		},
	}

	utils.Logger.LogF(100, "Sending request to LLM: %v", req)

	requestJson, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("error marshalling request: %v", err)
	}

	headers := map[string]string{
		"Authorization": "Bearer " + config.GetConfig().APIKey,
		"Content-Type":  "application/json",
	}

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		response, status, err := utils.DoParseReq[schema.GPTResponse](config.GetConfig().Url, requestJson, "POST", headers, config.HTTP_GPT_TIMEOUT)
		if err == nil && len(response.Choices) > 0 {
			err = processLLMResponse(alert, response.Choices[0].Message.Content)
			if err != nil {
				return fmt.Errorf("error processing LLM response: %v", err)
			}
			return nil
		}

		if status == 401 {
			return fmt.Errorf("invalid api-key")
		}
		lastErr = fmt.Errorf("attempt %d failed: %v (status: %d)", attempt, err, status)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	utils.Logger.LogF(500, "LLM appears to be DOWN - all %d attempts failed for alert %s. Provider: %s, URL: %s, Last error: %v",
		maxRetries, alert.ID, config.GetConfig().Provider, config.GetConfig().Url, lastErr)

	return fmt.Errorf("all attempts to call LLM failed: %v", lastErr)
}

func processLLMResponse(alert *schema.AlertFields, response string) error {
	response, err := clearJson(response)
	if err != nil {
		return fmt.Errorf("error clearing json: %v", err)
	}

	alertResponse, err := utils.ConvertFromJsonToStruct[schema.GPTAlertResponse](response)
	if err != nil {
		return fmt.Errorf("error converting json to struct: %v", err)
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

func clearJson(s string) (string, error) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 {
		return "", fmt.Errorf("no valid json found in gpt response")
	}

	return s[start : end+1], nil
}
