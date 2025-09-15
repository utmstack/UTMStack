package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
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
		return catcher.Error("sendRequestToOpenAI: alert is nil", nil, nil)
	}
	correlationContext, err := correlation.GetCorrelationContext(*alert)
	if err != nil {
		return catcher.Error("error getting correlation context", err, nil)
	}
	if correlationContext != "" {
		content = fmt.Sprintf("%s%s", content, correlationContext)
	}

	jsonContent, err := json.Marshal(alert)
	if err != nil {
		return catcher.Error("error marshalling alert", err, nil)
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

	catcher.Info("Sending request to LLM", map[string]any{"request": req})

	requestJson, err := json.Marshal(req)
	if err != nil {
		return catcher.Error("error marshalling request", err, nil)
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
				return catcher.Error("error processing LLM response", err, nil)
			}
			return nil
		}

		if status == 401 {
			return catcher.Error("invalid api-key", nil, nil)
		}
		lastErr = fmt.Errorf("attempt %d failed: %v (status: %d)", attempt, err, status)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	catcher.Error("LLM appears to be DOWN", lastErr, map[string]any{
		"attempts": maxRetries,
		"alert":    alert.ID,
		"provider": config.GetConfig().Provider,
		"url":      config.GetConfig().Url,
	})

	return catcher.Error("all attempts to call LLM failed", lastErr, map[string]any{})
}

func processLLMResponse(alert *schema.AlertFields, response string) error {
	response, err := clearJson(response)
	if err != nil {
		return catcher.Error("error clearing json", err, nil)
	}

	alertResponse, err := utils.ConvertFromJsonToStruct[schema.GPTAlertResponse](response)
	if err != nil {
		return catcher.Error("error converting json to struct", err, nil)
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
		return "", catcher.Error("no valid json found in gpt response", nil, nil)
	}

	return s[start : end+1], nil
}
