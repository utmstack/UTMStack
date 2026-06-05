package verifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// validateHTTPClient is the shared client used by the HTTP-based backend
// verifiers (o365, bitdefender). Ported from baseline.ValidateHTTPClient.
var validateHTTPClient = &http.Client{Timeout: 10 * time.Second}

const (
	o365GrantType     = "client_credentials"
	o365EndPointLogin = "/oauth2/v2.0/token"
)

type o365CloudEnvironment string

const (
	o365CloudCommercial o365CloudEnvironment = "Commercial"
	o365CloudGCC        o365CloudEnvironment = "GCC"
	o365CloudGCCHigh    o365CloudEnvironment = "GCCHigh"
	o365CloudDoD        o365CloudEnvironment = "DoD"
)

type o365CloudConfig struct {
	LoginAuthority     string
	ManagementEndpoint string
	Scope              string
}

var o365CloudConfigs = map[o365CloudEnvironment]o365CloudConfig{
	o365CloudCommercial: {
		LoginAuthority:     "https://login.microsoftonline.com/",
		ManagementEndpoint: "https://manage.office.com/",
		Scope:              "https://manage.office.com/.default",
	},
	o365CloudGCC: {
		LoginAuthority:     "https://login.microsoftonline.com/",
		ManagementEndpoint: "https://manage-gcc.office.com/",
		Scope:              "https://manage-gcc.office.com/.default",
	},
	o365CloudGCCHigh: {
		LoginAuthority:     "https://login.microsoftonline.us/",
		ManagementEndpoint: "https://manage.office365.us/",
		Scope:              "https://manage.office365.us/.default",
	},
	o365CloudDoD: {
		LoginAuthority:     "https://login.microsoftonline.us/",
		ManagementEndpoint: "https://manage.protection.apps.mil/",
		Scope:              "https://manage.protection.apps.mil/.default",
	},
}

func getO365CloudConfig(env o365CloudEnvironment) o365CloudConfig {
	if c, ok := o365CloudConfigs[env]; ok {
		return c
	}
	return o365CloudConfigs[o365CloudCommercial]
}

type microsoftLoginResponse struct {
	TokenType   string `json:"token_type,omitempty"`
	Expires     int    `json:"expires_in,omitempty"`
	ExtExpires  int    `json:"ext_expires_in,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

// verifyO365 — ported from modulekinds/o365/validate.go.
func verifyO365(c map[string]string) error {
	if err := required(c, "O365",
		"office365_client_id",
		"office365_client_secret",
		"office365_tenant_id",
	); err != nil {
		return err
	}

	clientId := c["office365_client_id"]
	clientSecret := c["office365_client_secret"]
	tenantId := c["office365_tenant_id"]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	env := o365CloudCommercial
	if v := c["office365_cloud_environment"]; v != "" {
		env = o365CloudEnvironment(v)
	}
	cc := getO365CloudConfig(env)

	requestUrl := fmt.Sprintf("%s%s%s", cc.LoginAuthority, tenantId, o365EndPointLogin)
	form := url.Values{}
	form.Set("grant_type", o365GrantType)
	form.Set("client_id", clientId)
	form.Set("client_secret", clientSecret)
	form.Set("scope", cc.Scope)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestUrl, bytes.NewBufferString(form.Encode()))
	if err != nil {
		return fmt.Errorf("Unable to validate Office 365 configuration. Please try again.")
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := validateHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot connect to Microsoft login service. Please verify your network connection.")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Received an invalid response from Microsoft. Please try again.")
	}

	var loginResp microsoftLoginResponse
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

	return validateO365ManagementAPIAccess(ctx, loginResp.TokenType, loginResp.AccessToken, cc.ManagementEndpoint, tenantId)
}

func validateO365ManagementAPIAccess(ctx context.Context, tokenType, accessToken, managementEndpoint, tenantId string) error {
	pingUrl := fmt.Sprintf("%sapi/v1.0/%s/activity/feed/subscriptions/list", managementEndpoint, tenantId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pingUrl, nil)
	if err != nil {
		return fmt.Errorf("Unable to validate Office 365 Management API access. Please try again.")
	}
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", tokenType, accessToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := validateHTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Cannot connect to the Office 365 Management API. Please verify your network connection.")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
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
