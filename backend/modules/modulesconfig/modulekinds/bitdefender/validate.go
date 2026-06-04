package bitdefender

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/utmstack/utmstack/backend/modules/modulesconfig/domain"
	"github.com/utmstack/utmstack/backend/modules/modulesconfig/modulekinds/baseline"
)

const endpointPush = "/v1.0/jsonrpc/push"

type bitdefenderRequest struct {
	JsonRPC string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	ID      string         `json:"id"`
	Params  map[string]any `json:"params"`
}

func (k *kind) ValidateConfiguration(ctx context.Context, _ *domain.UtmModule, configs []domain.UtmModuleGroupConfiguration) error {
	if err := baseline.RequireFields(configs, "Bitdefender",
		"connectionKey", "Connection Key",
		"accessUrl", "Access URL",
		"utmPublicIP", "Master IP",
		"companyIds", "Companies IDs",
	); err != nil {
		return err
	}

	connectionKey := baseline.ConfigValue(configs, "connectionKey")
	accessUrl := baseline.ConfigValue(configs, "accessUrl")

	if !strings.HasPrefix(accessUrl, "http://") && !strings.HasPrefix(accessUrl, "https://") {
		return fmt.Errorf("Access URL must start with http:// or https://")
	}

	body, err := json.Marshal(bitdefenderRequest{
		JsonRPC: "2.0",
		Method:  "getPushEventSettings",
		ID:      "1",
		Params:  map[string]any{},
	})
	if err != nil {
		return fmt.Errorf("Unable to validate Bitdefender configuration. Please try again.")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, accessUrl+endpointPush, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("Invalid Access URL '%s'. Please verify the Bitdefender Access URL format is correct.", accessUrl)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(connectionKey+":")))

	resp, err := baseline.ValidateHTTPClient.Do(req)
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errMsg, "timeout"), strings.Contains(errMsg, "deadline"):
			return fmt.Errorf("Connection to Bitdefender timed out. Please verify the Access URL '%s' is reachable from this server.", accessUrl)
		case strings.Contains(errMsg, "no such host"), strings.Contains(errMsg, "lookup"):
			return fmt.Errorf("Cannot resolve the Bitdefender Access URL '%s'. Please verify the hostname is correct.", accessUrl)
		case strings.Contains(errMsg, "connection refused"):
			return fmt.Errorf("Connection refused by Bitdefender at '%s'. Please verify the Access URL and port are correct.", accessUrl)
		}
		return fmt.Errorf("Cannot connect to Bitdefender at '%s'. Please verify the Access URL is correct and accessible from this server.", accessUrl)
	}
	defer resp.Body.Close()

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
					// -32000 with "Settings not set" means no push config yet — valid state
					if code == -32000 && strings.Contains(details, "Settings for event push service were not set") {
						return nil
					}
					if code == -32001 || code == -32002 {
						return fmt.Errorf("Invalid Connection Key. The Bitdefender API rejected the authentication. Please verify the Connection Key is correct.")
					}
					if code == -32601 {
						return fmt.Errorf("Bitdefender API method not found. Please verify the Access URL points to the correct GravityZone API version.")
					}
					if message, ok := errorMap["message"].(string); ok {
						combined := strings.ToLower(message + " " + details)
						if strings.Contains(combined, "unauthorized") ||
							strings.Contains(combined, "authentication") ||
							strings.Contains(combined, "invalid api key") ||
							strings.Contains(combined, "access denied") {
							return fmt.Errorf("Invalid Connection Key. Bitdefender returned: %s", message)
						}
						return fmt.Errorf("Bitdefender API error: %s", message)
					}
				}
			}
			return fmt.Errorf("Bitdefender API returned an error. Please verify your Connection Key and Access URL are correct.")
		}
		if _, hasResult := respBody["result"]; !hasResult && resp.StatusCode == 200 {
			if _, hasId := respBody["id"]; !hasId {
				return fmt.Errorf("Unexpected response from Bitdefender API. Please verify the Access URL points to a valid GravityZone API endpoint.")
			}
		}
	} else if resp.StatusCode == 200 {
		return fmt.Errorf("Unexpected response from Bitdefender API. The Access URL may not point to a valid GravityZone API endpoint.")
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return fmt.Errorf("Invalid Connection Key. Bitdefender rejected the authentication (HTTP 401).")
	case resp.StatusCode == http.StatusForbidden:
		return fmt.Errorf("Connection Key does not have permission to access the Push Event API (HTTP 403). Please verify the API Key permissions in GravityZone.")
	case resp.StatusCode == http.StatusNotFound:
		return fmt.Errorf("Bitdefender API endpoint not found (HTTP 404). Please verify the Access URL '%s' is correct.", accessUrl)
	case resp.StatusCode >= 500:
		return fmt.Errorf("Bitdefender server error (HTTP %d). The GravityZone service may be temporarily unavailable. Please try again later.", resp.StatusCode)
	case resp.StatusCode >= 400:
		return fmt.Errorf("Bitdefender API returned HTTP %d. Please verify your Connection Key and Access URL are correct.", resp.StatusCode)
	}
	return nil
}
