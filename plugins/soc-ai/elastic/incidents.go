package elastic

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/plugins/soc-ai/configurations"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

func CreateNewIncident(alertDetails schema.AlertGPTDetails) error {
	config := configurations.GetPluginConfig()
	url := config.Backend + configurations.API_INCIDENT_ENDPOINT
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": config.InternalKey,
	}

	t := time.Now()
	body := schema.CreateNewIncidentRequest{
		IncidentName:        fmt.Sprintf("INC-%s%s Incident in %s", t.Format("2006010215"), t.Format("04"), alertDetails.DataSource),
		IncidentDescription: fmt.Sprintf("AI GENERATED ANALYSIS: Multiple related alerts were detected and grouped in the %s datasource during the last 24 hours. Artificial intelligence classified this grouping as a possible incident", alertDetails.DataSource),
		IncidentAssignedTo:  "None",
		AlertList: schema.AlertList{{
			AlertID:       alertDetails.ID,
			AlertName:     alertDetails.Name,
			AlertStatus:   alertDetails.Status,
			AlertSeverity: alertDetails.Severity,
		}},
	}

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

func AddAlertToIncident(incidentId int, alertDetails schema.AlertGPTDetails) error {
	config := configurations.GetPluginConfig()
	url := config.Backend + configurations.API_INCIDENT_ADD_NEW_ALERT_ENDPOINT
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": config.InternalKey,
	}

	body := schema.AddNewAlertToIncidentRequest{
		IncidenId: incidentId,
		AlertList: schema.AlertList{{
			AlertID:       alertDetails.ID,
			AlertName:     alertDetails.Name,
			AlertStatus:   alertDetails.Status,
			AlertSeverity: alertDetails.Severity,
		}},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling body: %v", err)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, configurations.HTTP_TIMEOUT)
	if err != nil || (statusCode != http.StatusOK && statusCode != http.StatusCreated) {
		return fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	return nil
}

func GetIncidentsByPattern(pattern string) ([]schema.IncidentResp, error) {
	pattern = strings.ReplaceAll(pattern, " ", "%20")
	tnow := time.Now().UTC()
	t24hAfter := tnow.Add(24 * time.Hour)
	t24hBefore := tnow.Add(-24 * time.Hour)

	config := configurations.GetPluginConfig()
	url := config.Backend + configurations.API_INCIDENT_ENDPOINT + "?incidentName.contains=" + pattern + "&incidentCreatedDate.greaterThanOrEqual=" + t24hBefore.Format(time.RFC3339) + "&incidentCreatedDate.lessThanOrEqual=" + t24hAfter.Format(time.RFC3339) + "&incidentStatus.in=IN_REVIEW,OPEN&page=0&size=100"
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": config.InternalKey,
	}

	resp, statusCode, err := utils.DoReq(url, nil, "GET", headers, configurations.HTTP_TIMEOUT)
	if err != nil || statusCode != http.StatusOK {
		return nil, fmt.Errorf("error while doing request: %v, status: %d, response: %v", err, statusCode, string(resp))
	}

	var incidents []schema.IncidentResp
	err = json.Unmarshal(resp, &incidents)
	if err != nil {
		return nil, fmt.Errorf("error while unmarshalling response: %v", err)
	}

	return incidents, nil
}
