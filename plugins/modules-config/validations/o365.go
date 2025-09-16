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
	loginUrl      = "https://login.microsoftonline.com/"
	grantType     = "client_credentials"
	scope         = "https://manage.office.com/.default"
	endPointLogin = "/oauth2/v2.0/token"
)

type MicrosoftLoginResponse struct {
	TokenType   string `json:"token_type,omitempty"`
	Expires     int    `json:"expires_in,omitempty"`
	ExtExpires  int    `json:"ext_expires_in,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

func ValidateO365Config(config *config.ModuleGroup) error {
	var clientId, clientSecret, tenantId string

	if config == nil {
		return fmt.Errorf("O365 configuration is nil")
	}

	for _, cnf := range config.ModuleGroupConfigurations {
		switch cnf.ConfKey {
		case "office365_client_id":
			clientId = cnf.ConfValue
		case "office365_client_secret":
			clientSecret = cnf.ConfValue
		case "office365_tenant_id":
			tenantId = cnf.ConfValue
		}
	}

	if clientId == "" {
		return fmt.Errorf("Client ID is required in O365 configuration")
	}
	if clientSecret == "" {
		return fmt.Errorf("Client Secret is required in O365 configuration")
	}
	if tenantId == "" {
		return fmt.Errorf("Tenant ID is required in O365 configuration")
	}

	// Validate credentials by attempting to get an access token
	requestUrl := fmt.Sprintf("%s%s%s", loginUrl, tenantId, endPointLogin)

	data := url.Values{}
	data.Set("grant_type", grantType)
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("scope", scope)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("O365 authentication request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	var loginResp MicrosoftLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	if loginResp.Error != "" {
		return fmt.Errorf("O365 authentication failed: %s - %s", loginResp.Error, loginResp.ErrorDesc)
	}

	if loginResp.AccessToken == "" {
		return fmt.Errorf("O365 authentication failed: no access token received")
	}

	return nil
}
