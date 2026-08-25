package main

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

	"github.com/threatwinds/go-sdk/catcher"
)

// Template for the notification message sent to the tenant that flooded. It is
// formatted with the rule name, the number of alerts, the data source, and the
// window in hours.
const notificationMessageTemplate = "Correlation rule '%s' generated %d open, un-deduplicated alerts from data source '%s' in the last %dh and was automatically disabled to prevent alert flooding. If this volume is expected, use an Alert Tag Rule to mark it 'False positive', or add deduplicateBy/groupBy to the rule, then re-enable it."

// tenantHeader scopes every backend call. Without it the middleware treats an
// internal caller as tenantless and the backend falls back to the platform
// tenant, so the disable would land on the wrong tenant's rule list.
const tenantHeader = "X-Tenant-Id"

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

type ruleSearchResult struct {
	RelPath    string `json:"relPath"`
	Name       string `json:"name"`
	RuleActive bool   `json:"ruleActive"`
}

func (c *backendClient) resolveRule(ctx context.Context, tenantID, ruleName string) ([]ruleSearchResult, error) {
	endpoint := c.baseURL + "/api/v1/eventprocessing/correlation-rule/search-by-filters?ruleName=" + url.QueryEscape(ruleName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	req.Header.Set(tenantHeader, tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, catcher.Error("search-by-filters call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "ruleName": ruleName, "tenantId": tenantID,
		})
	}

	var candidates []ruleSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&candidates); err != nil {
		return nil, err
	}

	var matches []ruleSearchResult
	for i := range candidates {
		if strings.EqualFold(candidates[i].Name, ruleName) {
			matches = append(matches, candidates[i])
		}
	}
	if len(matches) > 1 {
		catcher.Warn("rule-flood-guard: ambiguous rule name collision, disabling every exact match", map[string]any{
			"ruleName": ruleName, "matches": len(matches), "tenantId": tenantID,
		})
	}
	return matches, nil
}

type activateDeactivateResponse struct {
	Changed bool `json:"changed"`
}

func (c *backendClient) Deactivate(ctx context.Context, tenantID, ruleName string) (bool, error) {
	matches, err := c.resolveRule(ctx, tenantID, ruleName)
	if err != nil {
		return false, err
	}
	if len(matches) == 0 {
		return false, nil
	}

	changed := false
	for i := range matches {
		if !matches[i].RuleActive {
			continue
		}
		ruleChanged, err := c.deactivateOne(ctx, tenantID, ruleName, matches[i].RelPath)
		if err != nil {
			return changed, err
		}
		if ruleChanged {
			changed = true
		}
	}
	catcher.Info("rule-flood-guard: exact-match rules processed for deactivation", map[string]any{
		"ruleName": ruleName, "matches": len(matches), "changed": changed, "tenantId": tenantID,
	})
	return changed, nil
}

func (c *backendClient) deactivateOne(ctx context.Context, tenantID, ruleName, relPath string) (bool, error) {
	endpoint := fmt.Sprintf("%s/api/v1/eventprocessing/correlation-rule/activate-deactivate?relPath=%s&active=false",
		c.baseURL, url.QueryEscape(relPath))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("X-Internal-Key", c.internalKey)
	req.Header.Set(tenantHeader, tenantID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return false, catcher.Error("activate-deactivate call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "ruleName": ruleName,
			"relPath": relPath, "tenantId": tenantID,
		})
	}

	var result activateDeactivateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}
	return result.Changed, nil
}

type notifyRequest struct {
	Source  string `json:"source"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *backendClient) Notify(ctx context.Context, tenantID, message string) error {
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
	req.Header.Set(tenantHeader, tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return catcher.Error("notify call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "tenantId": tenantID,
		})
	}
	return nil
}

func floodNotificationMessage(ruleName string, count int64, dataSource string, windowHours int) string {
	return fmt.Sprintf(notificationMessageTemplate, ruleName, count, dataSource, windowHours)
}
