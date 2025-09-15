package validations

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
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
		return catcher.Error("O365 configuration is nil", nil, nil)
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
		return catcher.Error("Client ID is required in O365 configuration", nil, nil)
	}
	if clientSecret == "" {
		return catcher.Error("Client Secret is required in O365 configuration", nil, nil)
	}
	if tenantId == "" {
		return catcher.Error("Tenant ID is required in O365 configuration", nil, nil)
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
		return catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return catcher.Error("O365 authentication request failed", err, nil)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return catcher.Error("failed to read response", err, nil)
	}

	var loginResp MicrosoftLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return catcher.Error("failed to parse response", err, nil)
	}

	if loginResp.Error != "" {
		return catcher.Error("O365 authentication failed", nil, map[string]any{
			"error":             loginResp.Error,
			"error_description": loginResp.ErrorDesc,
		})
	}

	if loginResp.AccessToken == "" {
		return catcher.Error("O365 authentication failed: no access token received", nil, nil)
	}

	return nil
}
