package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const notificationMessageTemplate = "Correlation rule '%s' generated over %d open, un-deduplicated alerts in the last 24h and was automatically disabled to prevent alert flooding. If this volume is expected, use an Alert Tag Rule to mark it 'False positive', or add deduplicateBy/groupBy to the rule, then re-enable it."

type backendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func newBackendClient(baseURL, internalKey string) *backendClient {
	return &backendClient{
		baseURL:     baseURL,
		internalKey: internalKey,
		httpClient:  &http.Client{Timeout: 15 * time.Second},
	}
}

type deactivateRequest struct {
	RuleName string `json:"ruleName"`
}

type deactivateResponse struct {
	Changed bool `json:"changed"`
}

func (c *backendClient) Deactivate(ctx context.Context, ruleName string) (bool, error) {
	payload, err := json.Marshal(deactivateRequest{RuleName: ruleName})
	if err != nil {
		return false, err
	}
	url := c.baseURL + "/api/v1/eventprocessing/internal/correlation-rule/deactivate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(payload))
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return false, catcher.Error("deactivate call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "ruleName": ruleName,
		})
	}

	var out deactivateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, err
	}
	return out.Changed, nil
}

type notifyRequest struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *backendClient) Notify(ctx context.Context, message string) error {
	payload, err := json.Marshal(notifyRequest{Source: "SYSTEM", Type: "WARNING", Message: message})
	if err != nil {
		return err
	}
	url := c.baseURL + "/api/v1/notifications"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("notify call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body),
		})
	}
	return nil
}

func floodNotificationMessage(ruleName string, threshold int64) string {
	return fmt.Sprintf(notificationMessageTemplate, ruleName, threshold)
}
