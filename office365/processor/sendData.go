package processor

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/quantfall/holmes"
	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/utils"
)

func SendToCorrelation(data []TransformedLog, h *holmes.Logger) error {
	h.Info("Uploading %d logs...", len(data))

	for _, log := range data {
		body, err := json.Marshal(log)
		if err != nil {
			h.Error("error encoding log to JSON: %v", err)
			continue
		}

		resp, status, err := utils.DoReq[map[string]interface{}](configuration.CORRELATIONURL, body, http.MethodPost, map[string]string{})
		if err != nil {
			h.Error(fmt.Sprintf("Error sending log to correlation engine: %v", err))
			continue
		} else if status != http.StatusOK && status != http.StatusCreated {
			h.Error(fmt.Sprintf("Error sending log to correlation engine with status code %d: %v", status, resp))
			continue
		}

		h.Info("Log successfully sent to correlation engine")
	}
	h.Info("All logs sent to correlation")
	return nil
}
