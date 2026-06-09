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
	"github.com/utmstack/UTMStack/plugins/feeds/config"
	"github.com/utmstack/UTMStack/plugins/feeds/internal/models"
)

const internalKeyHeader = "X-Internal-Key"

// ThreatWinds config keys in the appconfig store (utm_configuration_parameter).
const (
	twEnabledKey   = "utmstack.tw.enabled"
	twAPIKeyKey    = "utmstack.tw.apiKey"
	twAPISecretKey = "utmstack.tw.apiSecret"
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

// getJSON performs an internal-authenticated GET and decodes the body into out.
func (c *BackendClient) getJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return catcher.Error("failed to create request", err, nil)
	}
	req.Header.Set(internalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("request failed", err, nil)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("unexpected status from backend", nil, map[string]any{
			"url": url, "status": resp.StatusCode, "body": string(body),
		})
	}
	if out == nil {
		return nil
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return catcher.Error("failed to decode response", err, nil)
	}
	return nil
}

func (c *BackendClient) GetRecentIncidents(ctx context.Context) ([]*models.Incident, error) {
	var all []*models.Incident
	for _, status := range []string{"OPEN", "IN_REVIEW"} {
		url := fmt.Sprintf("%s/api/v1/incidents?incidentStatus=%s&sort=incidentCreatedDate,desc&size=100", c.baseURL, status)
		var batch []*models.Incident
		if err := c.getJSON(ctx, url, &batch); err != nil {
			return nil, err
		}
		all = append(all, batch...)
	}
	return all, nil
}

func (c *BackendClient) GetIncidentAlerts(ctx context.Context, incidentID int64) ([]*models.IncidentAlert, error) {
	url := fmt.Sprintf("%s/api/v1/incident-alerts?incidentId=%d&size=1000", c.baseURL, incidentID)
	var alerts []*models.IncidentAlert
	if err := c.getJSON(ctx, url, &alerts); err != nil {
		return nil, err
	}
	return alerts, nil
}

type ThreadWindsConfig struct {
	APIKey    string
	APISecret string
	Enabled   string
}

// getConfigValue reads a single appconfig parameter. Secrets are returned already
// decrypted by the per-key endpoint.
func (c *BackendClient) getConfigValue(ctx context.Context, key string) (string, error) {
	url := fmt.Sprintf("%s/api/v1/config/%s", c.baseURL, key)
	var resp struct {
		Value string `json:"value"`
	}
	if err := c.getJSON(ctx, url, &resp); err != nil {
		return "", err
	}
	return resp.Value, nil
}

func (c *BackendClient) GetThreadWindsConfig(ctx context.Context) (*ThreadWindsConfig, error) {
	enabled, err := c.getConfigValue(ctx, twEnabledKey)
	if err != nil {
		return nil, err
	}
	apiKey, err := c.getConfigValue(ctx, twAPIKeyKey)
	if err != nil {
		return nil, err
	}
	apiSecret, err := c.getConfigValue(ctx, twAPISecretKey)
	if err != nil {
		return nil, err
	}
	return &ThreadWindsConfig{Enabled: enabled, APIKey: apiKey, APISecret: apiSecret}, nil
}

// putConfigValue updates a single appconfig parameter (the key must be seeded;
// the backend encrypts it when the parameter is a secret).
func (c *BackendClient) putConfigValue(ctx context.Context, key, value string) error {
	url := fmt.Sprintf("%s/api/v1/config/%s", c.baseURL, key)
	payload, err := json.Marshal(map[string]string{"value": value})
	if err != nil {
		return catcher.Error("failed to marshal config value", err, nil)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return catcher.Error("failed to create request", err, nil)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("request failed", err, nil)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("unexpected status from backend", nil, map[string]any{
			"key": key, "status": resp.StatusCode, "body": string(body),
		})
	}
	return nil
}

func (c *BackendClient) SaveThreadWindsCredentials(ctx context.Context, apiKey, apiSecret string) error {
	if err := c.putConfigValue(ctx, twAPIKeyKey, apiKey); err != nil {
		return err
	}
	if err := c.putConfigValue(ctx, twAPISecretKey, apiSecret); err != nil {
		return err
	}
	catcher.Info("ThreadWinds credentials saved successfully", nil)
	return nil
}
