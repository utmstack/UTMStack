package updater

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/utmstack/UTMStack/installer/config"
	"github.com/utmstack/UTMStack/installer/utils"
)

// IsBackendMaintenanceError checks if the error is due to backend being unavailable (maintenance or down)
func IsBackendMaintenanceError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Check for common signs of backend being unavailable
	return (strings.Contains(errStr, "<!DOCTYPE html") || 
		strings.Contains(errStr, "<html>") ||
		strings.Contains(errStr, "invalid character '<'") ||
		strings.Contains(errStr, "UTMStack - Maintenance") ||
		strings.Contains(errStr, "currently under maintenance") ||
		strings.Contains(errStr, "502 Bad Gateway") ||
		strings.Contains(errStr, "503 Service Unavailable") ||
		strings.Contains(errStr, "504 Gateway Timeout")) ||
		(strings.Contains(errStr, "status code: 502") || 
		 strings.Contains(errStr, "status code: 503") ||
		 strings.Contains(errStr, "status code: 504"))
}

func getConfigFromBackend(id uint) ([]ConfigBackend, error) {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	resp, status, err := utils.DoReq[[]ConfigBackend](
		fmt.Sprintf(config.BackendConfigEndpoint, id),
		nil,
		http.MethodGet,
		map[string]string{
			"Utm-Internal-Key": config.GetConfig().InternalKey,
			"Content-Type":     "application/json",
		},
		transCfg,
	)
	if err != nil || status != http.StatusOK {
		return nil, fmt.Errorf("error getting config from backend: status code: %d, error: %v", status, err)
	}

	if len(resp) <= 0 {
		return nil, fmt.Errorf("error getting config from backend: empty response")
	}
	return resp, nil
}

func updateConfigInBackend(cnf []ConfigBackend, id uint) error {
	transCfg := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	jsonData, err := json.Marshal(cnf)
	if err != nil {
		return fmt.Errorf("error marshalling updated instance info: %v", err)
	}

	_, status, err := utils.DoReq[string](
		fmt.Sprintf(config.BackendConfigEndpoint, id),
		jsonData,
		http.MethodPut,
		map[string]string{
			"Utm-Internal-Key": config.GetConfig().InternalKey,
			"Content-Type":     "application/json",
		},
		transCfg,
	)
	if err != nil || status != http.StatusOK {
		return fmt.Errorf("error updating instance info in backend: status code: %d, error: %v", status, err)
	}

	return nil
}
