package gpt

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

type GPTClient struct{}

var (
	client     *GPTClient
	clientOnce sync.Once
)

func GetGPTClient() *GPTClient {
	clientOnce.Do(func() {
		client = &GPTClient{}
	})
	return client
}

func (c *GPTClient) Request(alert schema.AlertGPTDetails) (string, error) {
	content := configurations.GPT_INSTRUCTION

	if alert.Description != "" {
		correlationContext := strings.Split(alert.Description, "\nHistorical Context:")
		if len(correlationContext) > 1 {
			content = fmt.Sprintf("%s%s",
				content, fmt.Sprintf(configurations.CORRELATION_CONTEXT, correlationContext[1]))
		}
	}

	if alert.Logs == "" || alert.Logs == " " {
		content += content + ". " + configurations.GPT_FALSE_POSITIVE
	}

	if alert.Logs != "" && alert.Logs != " " {
		logs := strings.Split(alert.Logs, configurations.LOGS_SEPARATOR)
		if len(logs) > 0 {
			alert.Logs = logs[len(logs)-1]
		}
	}

	jsonContent, err := json.Marshal(alert)
	if err != nil {
		return "", fmt.Errorf("error marshalling alert: %v", err)
	}

	req := schema.GPTRequest{
		Model: configurations.GetGPTConfig().Model,
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

	requestJson, error := json.Marshal(req)
	if error != nil {
		return "", fmt.Errorf("error marshalling request: %v", error)
	}

	utils.Logger.Info("Complete GPT Request JSON: %s", string(requestJson))

	headers := map[string]string{
		"Authorization": "Bearer " + configurations.GetGPTConfig().APIKey,
		"Content-Type":  "application/json",
	}

	response, status, err := utils.DoParseReq[schema.GPTResponse](configurations.GPT_API_ENDPOINT, requestJson, "POST", headers, configurations.HTTP_GPT_TIMEOUT)
	if err != nil {
		if status == 401 {
			return "", fmt.Errorf("invalid api-key")
		}
		return "", fmt.Errorf("error making request to GPT: %v", err)
	}

	return response.Choices[0].Message.Content, nil
}

func (c *GPTClient) ProcessResponse(response string) (schema.GPTAlertResponse, error) {
	alertRespProcessed := schema.GPTAlertResponse{}

	err := json.Unmarshal([]byte(response), &alertRespProcessed)
	if err != nil {
		return schema.GPTAlertResponse{}, fmt.Errorf("error while unmarshalling response: %v", err)
	}

	return alertRespProcessed, nil
}
