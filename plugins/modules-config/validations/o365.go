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
	grantType     = "client_credentials"
	endPointLogin = "/oauth2/v2.0/token"
)

type CloudEnvironment string

const (
	CloudCommercial CloudEnvironment = "Commercial"
	CloudGCC        CloudEnvironment = "GCC"
	CloudGCCHigh    CloudEnvironment = "GCCHigh"
	CloudDoD        CloudEnvironment = "DoD"
)

type CloudConfig struct {
	LoginAuthority     string
	ManagementEndpoint string
	Scope              string
}

func getCloudConfig(env CloudEnvironment) CloudConfig {
	configs := map[CloudEnvironment]CloudConfig{
		CloudCommercial: {
			LoginAuthority:     "https://login.microsoftonline.com/",
			ManagementEndpoint: "https://manage.office.com/",
			Scope:              "https://manage.office.com/.default",
		},
		CloudGCC: {
			LoginAuthority:     "https://login.microsoftonline.com/",
			ManagementEndpoint: "https://manage-gcc.office.com/",
			Scope:              "https://manage-gcc.office.com/.default",
		},
		CloudGCCHigh: {
			LoginAuthority:     "https://login.microsoftonline.us/",
			ManagementEndpoint: "https://manage.office365.us/",
			Scope:              "https://manage.office365.us/.default",
		},
		CloudDoD: {
			LoginAuthority:     "https://login.microsoftonline.us/",
			ManagementEndpoint: "https://manage.protection.apps.mil/",
			Scope:              "https://manage.protection.apps.mil/.default",
		},
	}

	cloudConfig, exists := configs[env]
	if !exists {
		return configs[CloudCommercial]
	}
	return cloudConfig
}

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
	var cloudEnvironment CloudEnvironment = CloudCommercial

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
		case "office365_cloud_environment":
			if cnf.ConfValue != "" {
				cloudEnvironment = CloudEnvironment(cnf.ConfValue)
			}
		}
	}

	if clientId == "" {
		return fmt.Errorf("client ID is required in O365 configuration")
	}
	if clientSecret == "" {
		return fmt.Errorf("client secret is required in O365 configuration")
	}
	if tenantId == "" {
		return fmt.Errorf("tenant ID is required in O365 configuration")
	}

	cloudConfig := getCloudConfig(cloudEnvironment)

	// Validate credentials by attempting to get an access token
	requestUrl := fmt.Sprintf("%s%s%s", cloudConfig.LoginAuthority, tenantId, endPointLogin)

	data := url.Values{}
	data.Set("grant_type", grantType)
	data.Set("client_id", clientId)
	data.Set("client_secret", clientSecret)
	data.Set("scope", cloudConfig.Scope)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodPost, requestUrl, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("Unable to validate Office 365 configuration. Please try again.")
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot connect to Microsoft login service. Please verify your network connection.")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Received an invalid response from Microsoft. Please try again.")
	}

	var loginResp MicrosoftLoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("Received an unexpected response from Microsoft. Please try again.")
	}

	if loginResp.Error != "" {
		desc := strings.ToLower(loginResp.ErrorDesc)
		switch loginResp.Error {
		case "invalid_client":
			if strings.Contains(desc, "aadsts7000215") || strings.Contains(desc, "secret") {
				return fmt.Errorf("Invalid Client Secret. Please verify the Client Secret value (not the Secret ID) is correct and has not expired.")
			}
			return fmt.Errorf("Invalid Client ID or Client Secret. Please verify your Office 365 API credentials.")
		case "unauthorized_client":
			if strings.Contains(desc, "aadsts70001") || strings.Contains(desc, "not found") {
				return fmt.Errorf("Client ID was not found in the tenant '%s'. Please verify the Client ID and Tenant ID are correct.", tenantId)
			}
			return fmt.Errorf("The application is not authorized. Please verify the Client ID has the required API permissions in Azure AD.")
		case "invalid_grant":
			if strings.Contains(desc, "aadsts65001") || strings.Contains(desc, "consent") {
				return fmt.Errorf("Admin consent is required. Please grant admin consent for the Office 365 Management API permissions in Azure AD.")
			}
			return fmt.Errorf("Office 365 authentication failed. Please verify your Client ID, Client Secret, and Tenant ID.")
		case "invalid_request":
			if strings.Contains(desc, "tenant") {
				return fmt.Errorf("Invalid Tenant ID '%s'. Please verify the Office 365 Tenant ID is correct.", tenantId)
			}
			return fmt.Errorf("Invalid authentication request. Please verify your Client ID and Tenant ID are correct.")
		default:
			return fmt.Errorf("Office 365 authentication failed: %s. Please verify your Client ID, Client Secret, and Tenant ID.", loginResp.ErrorDesc)
		}
	}

	if loginResp.AccessToken == "" {
		return fmt.Errorf("Office 365 did not return an access token. Please verify your Client ID, Client Secret, and Tenant ID are correct.")
	}

	if err := validateManagementAPIAccess(loginResp.TokenType, loginResp.AccessToken, cloudConfig.ManagementEndpoint, tenantId); err != nil {
		return err
	}

	return nil
}

func validateManagementAPIAccess(tokenType, accessToken, managementEndpoint, tenantId string) error {
	pingUrl := fmt.Sprintf("%sapi/v1.0/%s/activity/feed/subscriptions/list",
		managementEndpoint,
		tenantId)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, pingUrl, nil)
	if err != nil {
		return fmt.Errorf("Unable to validate Office 365 Management API access. Please try again.")
	}

	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenType, accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot connect to the Office 365 Management API. Please verify your network connection.")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return fmt.Errorf("Office 365 Management API rejected the authentication (HTTP 401). The access token may be invalid. Please try saving the configuration again.")
	case http.StatusForbidden:
		return fmt.Errorf("The application does not have permission to access the Office 365 Management API (HTTP 403). Please add the 'ActivityFeed.Read' permission and grant admin consent in Azure AD.")
	case http.StatusNotFound:
		return fmt.Errorf("Office 365 Management API endpoint not found (HTTP 404). Please verify the Tenant ID '%s' and Cloud Environment are correct.", tenantId)
	default:
		return fmt.Errorf("Office 365 Management API returned HTTP %d. Please verify the app has the required permissions in Azure AD.", resp.StatusCode)
	}
}
