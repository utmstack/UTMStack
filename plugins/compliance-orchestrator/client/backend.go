package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

type BackendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewBackendClient() *BackendClient {
	raw := plugins.PluginCfg("com.utmstack", false).Get("backend").String()

	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "https://" + raw
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
