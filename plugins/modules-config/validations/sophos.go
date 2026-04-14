package validations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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
		return fmt.Errorf("Sophos configuration is not provided")
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
		return fmt.Errorf("Unable to validate Sophos configuration. Please try again.")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline") {
			return fmt.Errorf("Connection to Sophos timed out. Please verify your network can reach id.sophos.com.")
		}
		if strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup") {
			return fmt.Errorf("Cannot resolve Sophos authentication server. Please verify your DNS and network connectivity.")
		}
		return fmt.Errorf("Cannot connect to Sophos authentication service. Please verify your network connection can reach id.sophos.com.")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Received an invalid response from Sophos. Please try again.")
	}

	var response map[string]any
	if err := json.Unmarshal(body, &response); err != nil {
		return fmt.Errorf("Received an unexpected response from Sophos. Please try again.")
	}

	if resp.StatusCode != http.StatusOK {
		if errorCode, hasError := response["errorCode"]; hasError {
			switch errorCode {
			case "oauth.invalid_client_secret":
				return fmt.Errorf("Invalid Client Secret. Please verify the Sophos Client Secret is correct and has not been regenerated.")
			case "oauth.invalid_client_id":
				return fmt.Errorf("Invalid Client ID. Please verify the Sophos Client ID is correct. You can find it in Sophos Central under Global Settings > API Credentials.")
			case "oauth.client_disabled":
				return fmt.Errorf("The Sophos API Client is disabled. Please enable it in Sophos Central under Global Settings > API Credentials.")
			default:
				message := ""
				if msg, ok := response["message"].(string); ok {
					message = msg
				}
				return fmt.Errorf("Sophos authentication failed: %s. Please verify your Client ID and Client Secret.", message)
			}
		}
		if errorCode, hasError := response["error"]; hasError {
			errorDesc := ""
			if desc, ok := response["error_description"].(string); ok {
				errorDesc = desc
			}
			errStr := fmt.Sprintf("%v", errorCode)
			if strings.Contains(errStr, "invalid_client") {
				return fmt.Errorf("Invalid Client ID or Client Secret. Please verify your Sophos API credentials.")
			}
			return fmt.Errorf("Sophos authentication failed: %s. Please verify your Client ID and Client Secret.", errorDesc)
		}

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return fmt.Errorf("Invalid Client ID or Client Secret (HTTP 401). Please verify your Sophos API credentials.")
		case http.StatusForbidden:
			return fmt.Errorf("Sophos API access denied (HTTP 403). Please verify the API Client has the required permissions in Sophos Central.")
		case http.StatusTooManyRequests:
			return fmt.Errorf("Sophos API rate limit exceeded. Please wait a moment and try again.")
		default:
			return fmt.Errorf("Sophos authentication failed (HTTP %d). Please verify your Client ID and Client Secret.", resp.StatusCode)
		}
	}

	accessToken, ok := response["access_token"].(string)
	if !ok || accessToken == "" {
		return fmt.Errorf("Sophos did not return an access token. Please verify your Client ID and Client Secret are correct.")
	}

	return nil
}
