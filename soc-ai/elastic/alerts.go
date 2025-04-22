package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/utmstack/soc-ai/configurations"
	"github.com/utmstack/soc-ai/schema"
	"github.com/utmstack/soc-ai/utils"
)

func GetAlertsInfo(id string) (schema.Alert, error) {
	result, err := ElasticSearch(configurations.ALERT_INDEX_PATTERN, "id", id)
	if err != nil {
		return schema.Alert{}, fmt.Errorf("error while getting alert %s info: %v", id, err)
	}

	var alertDetails schema.AlertDetails
	err = json.Unmarshal(result, &alertDetails)
	if err != nil {
		return schema.Alert{}, fmt.Errorf("error decoding response: %v", err)
	}

	if len(alertDetails) == 0 {
		return schema.Alert{}, fmt.Errorf("no alert found for id")
	}

	if len(alertDetails[0].Logs) > 0 {
		var logs []string
		if len(alertDetails[0].Logs) > 3 {
			logs = alertDetails[0].Logs[:3]
		} else {
			logs = alertDetails[0].Logs
		}

		for i, log := range logs {
			resp, err := ElasticSearch(configurations.LOGS_INDEX_PATTERN, "id", log)
			if err != nil {
				continue
			}
			alertDetails[0].Logs[i] = string(resp)
		}
	}
	return alertDetails[0], nil
}

func ChangeAlertStatus(id string, status int, observations string) error {
	url := configurations.GetPanelServiceName() + configurations.API_ALERT_STATUS_ENDPOINT
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": configurations.GetInternalKey(),
	}

	body := schema.ChangeAlertStatus{AlertIDs: []string{id}, Status: status, StatusObservation: observations}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling body: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, configurations.HTTP_TIMEOUT)
	if err != nil || statusCode != http.StatusOK {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	return nil
}
