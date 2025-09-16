package validations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/utmstack/UTMStack/plugins/modules-config/config"
)

const (
	sophosAuthURL   = "https://id.sophos.com/api/v2/oauth2/token"
	sophosWhoamiURL = "https://api.central.sophos.com/whoami/v1"
)

func ValidateSophosConfig(config *config.ModuleGroup) error {
	var clientID, clientSecret string

	if config == nil {
		return fmt.Errorf("Sophos configuration is nil")
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
		return fmt.Errorf("Client ID is required in Sophos configuration")
	}
	if clientSecret == "" {
		return fmt.Errorf("Client Secret is required in Sophos configuration")
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("scope", "token")

	req, err := http.NewRequest(http.MethodPost, sophosAuthURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Sophos authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if errorCode, hasError := response["errorCode"]; hasError {
			message := ""
			if msg, ok := response["message"].(string); ok {
				message = msg
			}
			if errorCode == "oauth.invalid_client_secret" {
				return fmt.Errorf("Sophos authentication failed: Invalid Client Secret")
			}
			if errorCode == "oauth.invalid_client_id" {
				return fmt.Errorf("Sophos authentication failed: Invalid Client ID")
			}
			return fmt.Errorf("Sophos authentication failed: %v - %s", errorCode, message)
		}
		if errorCode, hasError := response["error"]; hasError {
			errorDesc := ""
			if desc, ok := response["error_description"].(string); ok {
				errorDesc = desc
			}
			return fmt.Errorf("Sophos authentication failed: %v - %s", errorCode, errorDesc)
		}
		return fmt.Errorf("Sophos authentication failed with status %d", resp.StatusCode)
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		var fields []string
		for k := range response {
			fields = append(fields, k)
		}
		return fmt.Errorf("Sophos authentication failed: no access token received. Response fields: %v", fields)
	}

	return nil
}
