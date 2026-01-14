package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/config"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/internal/models"
	"github.com/utmstack/UTMStack/threadwinds-ingestion/utils"
)

const (
	threadwindsSectionID = 6
)

type BackendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewBackendClient(cfg *config.TWConfig) *BackendClient {
	return &BackendClient{
		baseURL:     cfg.BackendURL,
		internalKey: cfg.InternalKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *BackendClient) GetRecentIncidents(ctx context.Context) ([]*models.Incident, error) {
	url := fmt.Sprintf("%s/api/utm-incidents?incidentStatus.in=OPEN,IN_REVIEW&sort=incidentCreatedDate,desc&size=100", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, catcher.Error("request failed", err, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, catcher.Error("unexpected status from backend", nil, map[string]any{
			"status": resp.StatusCode,
			"body":   string(body),
		})
	}

	var incidents []*models.Incident
	if err := json.NewDecoder(resp.Body).Decode(&incidents); err != nil {
		return nil, catcher.Error("failed to decode response", err, nil)
	}

	return incidents, nil
}

func (c *BackendClient) GetIncidentAlerts(ctx context.Context, incidentID int64) ([]*models.IncidentAlert, error) {
	url := fmt.Sprintf("%s/api/utm-incident-alerts?incidentId.equals=%d", c.baseURL, incidentID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, catcher.Error("request failed", err, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, catcher.Error("unexpected status from backend", nil, map[string]any{
			"status": resp.StatusCode,
			"body":   string(body),
		})
	}

	var alerts []*models.IncidentAlert
	if err := json.NewDecoder(resp.Body).Decode(&alerts); err != nil {
		return nil, catcher.Error("failed to decode response", err, nil)
	}

	return alerts, nil
}

type ThreadWindsConfig struct {
	APIKey    string
	APISecret string
	Enabled   string
	KeyID     int64
	SecretID  int64
}

type ConfigParameter struct {
	ID                   int64  `json:"id"`
	SectionID            int64  `json:"sectionId"`
	ConfParamShort       string `json:"confParamShort"`
	ConfParamLarge       string `json:"confParamLarge,omitempty"`
	ConfParamDescription string `json:"confParamDescription,omitempty"`
	ConfParamValue       string `json:"confParamValue"`
	ConfParamRequired    bool   `json:"confParamRequired,omitempty"`
	ConfParamDatatype    string `json:"confParamDatatype,omitempty"`
}

func (c *BackendClient) GetThreadWindsConfig(ctx context.Context) (*ThreadWindsConfig, error) {
	url := fmt.Sprintf("%s/api/utm-configuration-parameters?sectionId.equals=%d&size=100", c.baseURL, threadwindsSectionID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, catcher.Error("request failed", err, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, catcher.Error("unexpected status from backend", nil, map[string]any{
			"status": resp.StatusCode,
			"body":   string(body),
		})
	}

	var params []ConfigParameter
	if err := json.NewDecoder(resp.Body).Decode(&params); err != nil {
		return nil, catcher.Error("failed to decode response", err, nil)
	}

	config := &ThreadWindsConfig{}

	for _, param := range params {
		switch param.ConfParamShort {
		case "utmstack.tw.enabled":
			config.Enabled = param.ConfParamValue
		case "utmstack.tw.apiKey":
			config.APIKey = param.ConfParamValue
			config.KeyID = param.ID
		case "utmstack.tw.apiSecret":
			if param.ConfParamDatatype == "password" && param.ConfParamValue != "" {
				decrypted, err := utils.DecryptValue(param.ConfParamValue)
				if err != nil {
					return nil, catcher.Error("failed to decrypt API Secret", err, nil)
				}
				config.APISecret = decrypted
			} else {
				config.APISecret = param.ConfParamValue
			}
			config.SecretID = param.ID
		}
	}

	return config, nil
}

func (c *BackendClient) SaveThreadWindsCredentials(ctx context.Context, apiKey, apiSecret string, keyID, secretID int64) error {
	const threadwindsSectionID = 6
	url := fmt.Sprintf("%s/api/utm-configuration-parameters", c.baseURL)

	params := []ConfigParameter{
		{
			ID:                   keyID,
			SectionID:            threadwindsSectionID,
			ConfParamShort:       "utmstack.tw.apiKey",
			ConfParamLarge:       "ThreatWinds API Key",
			ConfParamDescription: "API Key for ThreatWinds integration.",
			ConfParamValue:       apiKey,
			ConfParamRequired:    true,
			ConfParamDatatype:    "text",
		},
		{
			ID:                   secretID,
			SectionID:            threadwindsSectionID,
			ConfParamShort:       "utmstack.tw.apiSecret",
			ConfParamLarge:       "ThreatWinds API Secret",
			ConfParamDescription: "API Secret for ThreatWinds integration.",
			ConfParamValue:       apiSecret,
			ConfParamRequired:    true,
			ConfParamDatatype:    "password",
		},
	}

	payload, err := json.Marshal(params)
	if err != nil {
		return catcher.Error("failed to marshal parameters", err, nil)
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(payload))
	if err != nil {
		return catcher.Error("failed to create request", err, nil)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("request failed", err, nil)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("unexpected status from backend", nil, map[string]any{
			"status": resp.StatusCode,
			"body":   string(body),
		})
	}

	catcher.Info("ThreadWinds credentials saved successfully", nil)
	return nil
}
