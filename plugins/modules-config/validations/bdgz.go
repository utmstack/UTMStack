package validations

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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
		return fmt.Errorf("Bitdefender configuration is nil")
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
		return fmt.Errorf("Connection Key is required in Bitdefender configuration")
	}
	if accessUrl == "" {
		return fmt.Errorf("Access URL is required in Bitdefender configuration")
	}
	if masterIp == "" {
		return fmt.Errorf("Master IP is required in Bitdefender configuration")
	}
	if companiesIDs == "" {
		return fmt.Errorf("Companies IDs is required in Bitdefender configuration")
	}

	if !strings.HasPrefix(accessUrl, "http://") && !strings.HasPrefix(accessUrl, "https://") {
		return fmt.Errorf("Access URL must start with http:// or https://")
	}

	authCode := generateAuthCode(connectionKey)

	testRequest := BitdefenderRequest{
		JsonRPC: "2.0",
		Method:  "getPushEventSettings",
		ID:      "1",
		Params:  map[string]interface{}{},
	}

	body, err := json.Marshal(testRequest)
	if err != nil {
		return fmt.Errorf("failed to create test request: %w", err)
	}

	req, err := http.NewRequest("POST", accessUrl+endpointPush, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", authCode)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Bitdefender API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, _ := io.ReadAll(resp.Body)

	// Log response for debugging (you can remove this later)
	if len(bodyBytes) > 0 {
		fmt.Printf("Response body: %s\n", string(bodyBytes))
	}

	var respBody map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &respBody); err == nil {
		if errorField, ok := respBody["error"]; ok {
			if errorMap, ok := errorField.(map[string]interface{}); ok {
				if code, ok := errorMap["code"].(float64); ok {
					details := ""
					if dataMap, ok := errorMap["data"].(map[string]interface{}); ok {
						if d, ok := dataMap["details"].(string); ok {
							details = d
						}
					}

					if code == -32000 && strings.Contains(details, "Settings for event push service were not set") {
						return nil
					}

					if code == -32001 || code == -32002 {
						return fmt.Errorf("Bitdefender authentication failed: invalid Connection Key")
					}

					if message, ok := errorMap["message"].(string); ok {
						combinedError := strings.ToLower(message + " " + details)
						if strings.Contains(combinedError, "unauthorized") ||
							strings.Contains(combinedError, "authentication") ||
							strings.Contains(combinedError, "invalid api key") ||
							strings.Contains(combinedError, "access denied") {
							return fmt.Errorf("Bitdefender authentication failed: %s", message)
						}
					}
				}
				if message, ok := errorMap["message"].(string); ok {
					return fmt.Errorf("Bitdefender API error: %s", message)
				}
			}
			return fmt.Errorf("Bitdefender API error: %v", errorField)
		}

		if _, hasResult := respBody["result"]; !hasResult && resp.StatusCode == 200 {
			if _, hasId := respBody["id"]; !hasId {
				return fmt.Errorf("Invalid response format from Bitdefender API")
			}
		}
	} else if resp.StatusCode == 200 {
		return fmt.Errorf("Invalid JSON response from Bitdefender API")
	}

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("Bitdefender authentication failed: invalid Connection Key")
	}

	if resp.StatusCode == 404 {
		return fmt.Errorf("Bitdefender API endpoint not found. Please check the Access URL")
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("Bitdefender API returned error status: %d", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("Unexpected response status: %d", resp.StatusCode)
	}

	return nil
}

func generateAuthCode(apiKey string) string {
	loginString := apiKey + ":"
	encodedBytes := base64.StdEncoding.EncodeToString([]byte(loginString))
	return "Basic " + encodedBytes
}
