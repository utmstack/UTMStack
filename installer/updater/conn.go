package updater

import (
	"fmt"
	"net/http"
	"time"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

var messagePrinted = false

func MonitorConnection(server string, timeSleep time.Duration, maxAttempts int, hasConnection *bool) {
	var consecutiveErrors int

	for {
		_, status, err := utils.DoReq[string](server+config.HealthEndpoint, nil, http.MethodGet, nil, nil)

		if err == nil && status == http.StatusOK {
			consecutiveErrors = 0
			*hasConnection = true
		} else {
			consecutiveErrors++
			if consecutiveErrors >= maxAttempts {
				if !messagePrinted {
					fmt.Println("\nNo Internet connection detected. Installation will continue offline.")
					messagePrinted = true
				}
				*hasConnection = false
			}
		}

		time.Sleep(timeSleep)
	}
}
