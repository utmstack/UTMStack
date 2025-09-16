package elastic

import (
	"encoding/json"
	"net/http"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/soc-ai/config"
	"github.com/utmstack/UTMStack/plugins/soc-ai/schema"
	"github.com/utmstack/UTMStack/plugins/soc-ai/utils"
)

func ChangeAlertStatus(id string, status int, dataSource string, observations string) error {
	url := config.GetConfig().Backend + config.API_ALERT_STATUS_ENDPOINT
	headers := map[string]string{
		"Content-Type":     "application/json",
		"Utm-Internal-Key": config.GetConfig().InternalKey,
	}

	body := schema.ChangeAlertStatus{AlertIDs: []string{id}, Status: status, DataSource: dataSource, StatusObservation: observations}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return catcher.Error("error marshalling body", err, nil)
	}

	resp, statusCode, err := utils.DoReq(url, bodyBytes, "POST", headers, config.HTTP_TIMEOUT)
	if err != nil || statusCode != http.StatusOK {
		return catcher.Error("error while doing request", err, map[string]any{"status": statusCode, "response": string(resp)})
	}

	catcher.Info("Alert status changed successfully", map[string]any{"alert_id": id})

	return nil
}
