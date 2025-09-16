package validations

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

const (
	endpointPush = "/v1.0/jsonrpc/push"
)

type BitdefenderRequest struct {
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	ID      string                 `json:"id"`
	Params  map[string]interface{} `json:"params"`
}

func ValidateBdgzConfig(config *config.ModuleGroup) error {
	var connectionKey, accessUrl, masterIp, companiesIDs string

	if config == nil {
		return catcher.Error("Bitdefender configuration is nil", nil, nil)
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "connectionKey":
			connectionKey = cnf.ConfValue
		case "accessUrl":
			accessUrl = cnf.ConfValue
		case "utmPublicIP":
			masterIp = cnf.ConfValue
		case "companyIds":
			companiesIDs = cnf.ConfValue
		}
	}

	if connectionKey == "" {
		return catcher.Error("Connection Key is required in Bitdefender configuration", nil, nil)
	}
	if accessUrl == "" {
		return catcher.Error("Access URL is required in Bitdefender configuration", nil, nil)
	}
	if masterIp == "" {
		return catcher.Error("Master IP is required in Bitdefender configuration", nil, nil)
	}
	if companiesIDs == "" {
		return catcher.Error("Companies IDs is required in Bitdefender configuration", nil, nil)
	}

	if !strings.HasPrefix(accessUrl, "http://") && !strings.HasPrefix(accessUrl, "https://") {
		return catcher.Error("Access URL must start with http:// or https://", nil, nil)
	}

	authCode := generateAuthCode(connectionKey)

	testRequest := BitdefenderRequest{
		JsonRPC: "2.0",
		Method:  "getPushEventSettings",
		ID:      "1",
		Params:  map[string]any{},
	}

	body, err := json.Marshal(testRequest)
	if err != nil {
		return catcher.Error("failed to create test request", err, nil)
	}

	req, err := http.NewRequest("POST", accessUrl+endpointPush, bytes.NewBuffer(body))
	if err != nil {
		return catcher.Error("failed to create HTTP request", err, nil)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authCode)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return catcher.Error("Bitdefender API request failed", err, nil)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, _ := io.ReadAll(resp.Body)

	var respBody map[string]any
	if err := json.Unmarshal(bodyBytes, &respBody); err == nil {
		if errorField, ok := respBody["error"]; ok {
			if errorMap, ok := errorField.(map[string]any); ok {
				if code, ok := errorMap["code"].(float64); ok {
					details := ""
					if dataMap, ok := errorMap["data"].(map[string]any); ok {
						if d, ok := dataMap["details"].(string); ok {
							details = d
						}
					}

					if code == -32000 && strings.Contains(details, "Settings for event push service were not set") {
						return nil
					}

					if code == -32001 || code == -32002 {
						return catcher.Error("Bitdefender authentication failed: invalid Connection Key", nil, nil)
					}

					if message, ok := errorMap["message"].(string); ok {
						combinedError := strings.ToLower(message + " " + details)
						if strings.Contains(combinedError, "unauthorized") ||
							strings.Contains(combinedError, "authentication") ||
							strings.Contains(combinedError, "invalid api key") ||
							strings.Contains(combinedError, "access denied") {
							return catcher.Error("Bitdefender authentication failed", nil, map[string]any{"error": message})
						}
					}
				}
				if message, ok := errorMap["message"].(string); ok {
					return catcher.Error("Bitdefender API error", nil, map[string]any{"error": message})
				}
			}
			return catcher.Error("Bitdefender API error", nil, map[string]any{"error": errorField})
		}

		if _, hasResult := respBody["result"]; !hasResult && resp.StatusCode == 200 {
			if _, hasId := respBody["id"]; !hasId {
				return catcher.Error("Invalid response format from Bitdefender API", nil, nil)
			}
		}
	} else if resp.StatusCode == 200 {
		return catcher.Error("Invalid JSON response from Bitdefender API", nil, nil)
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return catcher.Error("Bitdefender authentication failed: invalid Connection Key", nil, nil)
	}

	if resp.StatusCode == 404 {
		return catcher.Error("Bitdefender API endpoint not found. Please check the Access URL", nil, nil)
	}

	if resp.StatusCode >= 400 {
		return catcher.Error("Bitdefender API returned error status", nil, map[string]any{"status_code": resp.StatusCode})
	}

	if resp.StatusCode != 200 {
		return catcher.Error("Unexpected response status", nil, map[string]any{"status_code": resp.StatusCode})
	}

	return nil
}

func generateAuthCode(apiKey string) string {
	loginString := apiKey + ":"
	encodedBytes := base64.StdEncoding.EncodeToString([]byte(loginString))
	return "Basic " + encodedBytes
}
