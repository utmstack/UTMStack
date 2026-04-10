package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/compliance-orchestrator/models"
)

type BackendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewBackendClient() *BackendClient {
	raw := plugins.PluginCfg("com.utmstack", false).Get("backend").String()

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw
	}

	return &BackendClient{
		baseURL:     raw,
		internalKey: plugins.PluginCfg("com.utmstack", false).Get("internalKey").String(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *BackendClient) HealthCheck(ctx context.Context) error {
	url := fmt.Sprintf("%s/api/ping", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return catcher.Error("failed to create backend ping request", err, nil)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("backend ping request failed", err, nil)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			catcher.Error("failed to close backend ping response body", err, nil)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return catcher.Error("backend ping returned non-200", nil, map[string]any{
			"status": resp.StatusCode,
		})
	}

	return nil
}

func (c *BackendClient) GetControlConfigs(ctx context.Context) ([]models.ControlConfig, error) {
	url := fmt.Sprintf("%s/api/compliance/control-config?page=0&size=1000&sort=id,asc", c.baseURL)
	var controls []models.ControlConfig

	var body, err = c.GetRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &controls); err != nil {
		return nil, err
	}
	return controls, nil
}

func (c *BackendClient) GetActiveIndexPatterns(ctx context.Context) ([]models.IndexPattern, error) {
	url := fmt.Sprintf("%s/api/utm-index-patterns?page=0&size=1000&sort=id,asc&isActive.equals=true", c.baseURL)
	var activeIndex []models.IndexPattern

	var body, err = c.GetRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, &activeIndex); err != nil {
		return nil, err
	}
	return activeIndex, nil
}

func (c *BackendClient) GetRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("utm-internal-key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("backend returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (b *BackendClient) IndexEvaluationResult(ctx context.Context, index string, doc any) error {
	osCfg := LoadOpenSearchConfig()
	client := NewOpenSearchHTTPClient()

	endpoint := fmt.Sprintf("%s/%s/_doc", osCfg.URL, index)

	jsonBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal evaluation result: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create index request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(osCfg.User, osCfg.Pass)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("index request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("indexing failed with status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
