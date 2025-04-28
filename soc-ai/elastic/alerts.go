package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

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

type AlertCorrelation struct {
	CurrentAlert    schema.Alert
	RelatedAlerts   []schema.Alert
	Classifications []string
}

func GetRelatedAlerts() ([]schema.GPTAlertResponse, error) {
	result, err := ElasticSearch(configurations.ALERT_INDEX_PATTERN, "*", "*")
	if err != nil {
		return nil, fmt.Errorf("error getting historical alerts: %v", err)
	}

	var alerts []schema.GPTAlertResponse
	err = json.Unmarshal(result, &alerts)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling alerts: %v", err)
	}

	return alerts, nil
}

func FindRelatedAlerts(currentAlert schema.Alert) (*AlertCorrelation, error) {
	correlation := &AlertCorrelation{
		CurrentAlert:    currentAlert,
		RelatedAlerts:   []schema.Alert{},
		Classifications: []string{},
	}

	historicalResponses, err := GetRelatedAlerts()
	if err != nil {
		return nil, err
	}

	var alertIDs []string
	for _, resp := range historicalResponses {
		alertIDs = append(alertIDs, resp.ActivityID)
	}

	for _, id := range alertIDs {
		alert, err := GetAlertsInfo(id)
		if err != nil {
			continue
		}

		if isAlertRelated(currentAlert, alert) {
			correlation.RelatedAlerts = append(correlation.RelatedAlerts, alert)

			for _, resp := range historicalResponses {
				if resp.ActivityID == alert.ID {
					correlation.Classifications = append(correlation.Classifications, resp.Classification)
					break
				}
			}
		}
	}

	return correlation, nil
}

func isAlertRelated(current, historical schema.Alert) bool {
	if current.Destination.IP != "" && current.Destination.IP == historical.Destination.IP {
		return true
	}
	if current.Destination.Port != 0 && current.Destination.Port == historical.Destination.Port {
		return true
	}
	if current.Destination.Host != "" && current.Destination.Host == historical.Destination.Host {
		return true
	}
	if current.Destination.User != "" && current.Destination.User == historical.Destination.User {
		return true
	}

	if current.Source.IP != "" && current.Source.IP == historical.Source.IP {
		return true
	}
	if current.Source.Port != 0 && current.Source.Port == historical.Source.Port {
		return true
	}
	if current.Source.Host != "" && current.Source.Host == historical.Source.Host {
		return true
	}
	if current.Source.User != "" && current.Source.User == historical.Source.User {
		return true
	}

	return false
}

func BuildCorrelationContext(correlation *AlertCorrelation) string {
	var context strings.Builder

	context.WriteString("\nHistorical Context:\n")
	context.WriteString(fmt.Sprintf("Found %d related alerts with similar characteristics:\n", len(correlation.RelatedAlerts)))

	for i, alert := range correlation.RelatedAlerts {
		context.WriteString(fmt.Sprintf("\nRelated Alert %d:\n", i+1))
		context.WriteString(fmt.Sprintf("- Name: %s\n", alert.Name))
		context.WriteString(fmt.Sprintf("- Severity: %s\n", alert.SeverityLabel))
		context.WriteString(fmt.Sprintf("- Category: %s\n", alert.Category))
		context.WriteString(fmt.Sprintf("- Classification: %s\n", correlation.Classifications[i]))
		context.WriteString(fmt.Sprintf("- Time: %s\n", alert.Timestamp))

		if alert.Source.IP != "" {
			context.WriteString(fmt.Sprintf("- Source IP: %s\n", alert.Source.IP))
		}
		if alert.Destination.IP != "" {
			context.WriteString(fmt.Sprintf("- Destination IP: %s\n", alert.Destination.IP))
		}
		if alert.Source.Host != "" {
			context.WriteString(fmt.Sprintf("- Source Host: %s\n", alert.Source.Host))
		}
		if alert.Destination.Host != "" {
			context.WriteString(fmt.Sprintf("- Destination Host: %s\n", alert.Destination.Host))
		}
		if alert.Source.User != "" {
			context.WriteString(fmt.Sprintf("- Source User: %s\n", alert.Source.User))
		}
		if alert.Destination.User != "" {
			context.WriteString(fmt.Sprintf("- Destination User: %s\n", alert.Destination.User))
		}
		if alert.Source.Port != 0 {
			context.WriteString(fmt.Sprintf("- Source Port: %d\n", alert.Source.Port))
		}
		if alert.Destination.Port != 0 {
			context.WriteString(fmt.Sprintf("- Destination Port: %d\n", alert.Destination.Port))
		}
		if alert.Protocol != "" {
			context.WriteString(fmt.Sprintf("- Protocol: %s\n", alert.Protocol))
		}
		if alert.Severity != 0 {
			context.WriteString(fmt.Sprintf("- Severity: %d\n", alert.Severity))
		}
	}

	return context.String()
}
