package processor

import (
	"fmt"
	"strings"
	"time"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/gpt"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

func (p *Processor) processGPTRequests() {
	for alert := range p.GPTQueue {
		response, err := gpt.GetGPTClient().Request(alert)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while making request to GPT: %v", err), alert.AlertID)
			continue
		}

		response, err = clearJson(response)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while cleaning json response: %v", err), alert.AlertID)
			continue
		}

		alertResponse, err := utils.ConvertFromJsonToStruct[schema.GPTAlertResponse](response)
		if err != nil {
			p.RegisterError(fmt.Sprintf("error while converting response to struct: %v", err), alert.AlertID)
			continue
		}

		nextSteps := []string{}
		for _, step := range alertResponse.NextSteps {
			nextSteps = append(nextSteps, fmt.Sprintf("%s:: %s", step.Action, step.Details))
		}

		alert.GPTTimestamp = time.Now().UTC().Format("2006-01-02T15:04:05.999999Z07:00")
		alert.GPTClassification = alertResponse.Classification
		alert.GPTReasoning = strings.Join(alertResponse.Reasoning, configurations.LOGS_SEPARATOR)
		alert.GPTNextSteps = strings.Join(nextSteps, "\n")

		p.ElasticQueue <- alert
	}
}

func clearJson(s string) (string, error) {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")

	if start == -1 || end == -1 {
		return "", fmt.Errorf("no valid json found in gpt response")
	}

	return s[start : end+1], nil
}
