package processor

import (
	"encoding/json"
	"net/http"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/aws/configuration"
	"github.com/utmstack/UTMStack/aws/utils"
)

func SendToCorrelation(data []TransformedLog) *logger.Error {
	for _, log := range data {
		body, err := json.Marshal(log)
		if err != nil {
			utils.Logger.ErrorF("error encoding log to JSON: %v", err)
			continue
		}

		_, status, e := utils.DoReq[map[string]interface{}](configuration.CORRELATIONURL, body, http.MethodPost, map[string]string{})
		if e != nil {
			utils.Logger.ErrorF("error sending log to correlation engine: %v", e)
			continue
		} else if status != http.StatusOK && status != http.StatusCreated {
			utils.Logger.ErrorF("error sending log to correlation engine: status %v", status)
			continue
		}

	}

	return nil
}
