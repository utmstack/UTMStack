package updater

import (
	"net/http"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/license/utils"
)

func MonitorConnection(server string, timeSleep time.Duration, maxAttempts int, hasConnection *bool) {
	var consecutiveErrors int

	for {
		_, status, err := utils.DoReq[string](server+config.HealthEndpoint, nil, http.MethodGet, nil)

		if err == nil && status == http.StatusOK {
			consecutiveErrors = 0
			*hasConnection = true
		} else {
			consecutiveErrors++
			if consecutiveErrors >= maxAttempts {
				*hasConnection = false
			}
		}

		time.Sleep(timeSleep)
	}
}
