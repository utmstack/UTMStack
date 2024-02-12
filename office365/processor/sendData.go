package processor

import (
	"encoding/json"
	"net/http"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/office365/configuration"
	"github.com/utmstack/UTMStack/office365/utils"
)

func SendToCorrelation(data []TransformedLog) *logger.Error {
	utils.Logger.Info("uploading %d logs...", len(data))

	for _, log := range data {
		body, err := json.Marshal(log)
		if err != nil {
			utils.Logger.ErrorF(http.StatusInternalServerError, "error encoding log to JSON: %v", err)
			continue
		}

		_, status, e := utils.DoReq[map[string]interface{}](configuration.CORRELATIONURL, body, http.MethodPost, map[string]string{})
		if e != nil {
			continue
		} else if status != http.StatusOK && status != http.StatusCreated {
			continue
		}

		utils.Logger.Info("log successfully sent to correlation engine")
	}

	utils.Logger.Info("all logs were sent to correlation")
	return nil
}
