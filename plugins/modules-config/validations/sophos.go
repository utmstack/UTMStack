package validations

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

const (
	sophosAuthURL   = "https://id.sophos.com/api/v2/oauth2/token"
	sophosWhoamiURL = "https://api.central.sophos.com/whoami/v1"
)

func ValidateSophosConfig(config *config.ModuleGroup) error {
	var clientID, clientSecret string

	if config == nil {
		return catcher.Error("Sophos configuration is nil", nil, nil)
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "sophos_client_id":
			clientID = cnf.ConfValue
		case "sophos_x_api_key":
			clientSecret = cnf.ConfValue
		}
	}

	if clientID == "" {
		return catcher.Error("Client ID is required in Sophos configuration", nil, nil)
	}
	if clientSecret == "" {
		return catcher.Error("Client Secret is required in Sophos configuration", nil, nil)
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "token")

	req, err := http.NewRequest(http.MethodPost, sophosAuthURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return catcher.Error("Sophos authentication request failed", err, nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return catcher.Error("failed to read response", err, nil)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return catcher.Error("failed to parse response", err, nil)
	}

	if resp.StatusCode != http.StatusOK {
		if errorCode, hasError := response["errorCode"]; hasError {
			message := ""
			if msg, ok := response["message"].(string); ok {
				message = msg
			}
			if errorCode == "oauth.invalid_client_secret" {
				return catcher.Error("Sophos authentication failed: Invalid Client Secret", nil, nil)
			}
			if errorCode == "oauth.invalid_client_id" {
				return catcher.Error("Sophos authentication failed: Invalid Client ID", nil, nil)
			}
			return catcher.Error("Sophos authentication failed", nil, map[string]any{
				"error_code": errorCode,
				"message":    message,
			})
		}
		if errorCode, hasError := response["error"]; hasError {
			errorDesc := ""
			if desc, ok := response["error_description"].(string); ok {
				errorDesc = desc
			}
			return catcher.Error("Sophos authentication failed", nil, map[string]any{
				"error_code": errorCode,
				"message":    errorDesc,
			})
		}
		return catcher.Error("Sophos authentication failed", nil, map[string]any{
			"status_code": resp.StatusCode,
		})
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		var fields []string
		for k := range response {
			fields = append(fields, k)
		}
		return catcher.Error("Sophos authentication failed", nil, map[string]any{
			"response_fields": fields,
		})
	}

	return nil
}
