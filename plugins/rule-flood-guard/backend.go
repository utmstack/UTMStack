package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

// The window is interpolated rather than hardcoded: windowHours is
// configurable, so a fixed "24h" would misreport the period the count covers.
const notificationMessageTemplate = "Correlation rule '%s' generated %d open, un-deduplicated alerts from data source '%s' for tenant '%s' in the last %dh and was automatically disabled for that tenant to prevent alert flooding. If this volume is expected, use an Alert Tag Rule to mark it 'False positive', or add deduplicateBy/groupBy to the rule, then re-enable it."

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

// platformTenant is the operator's tenant. It gets a copy of every flood
// notification so the operator keeps the instance-wide visibility they had
// before the disable became per-tenant.
const platformTenant = "ce66672c-e36d-4761-a8c8-90058fee1a24"

func (c *backendClient) Notify(ctx context.Context, tenantID, message string) error {
	err := c.notifyTenant(ctx, tenantID, message)
	if tenantID != platformTenant {
		err = errors.Join(err, c.notifyTenant(ctx, platformTenant, message))
	}
	return err
}

func (c *backendClient) notifyTenant(ctx context.Context, tenantID, message string) error {
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

func floodNotificationMessage(tenantID, ruleName string, count int64, dataSource string, windowHours int) string {
	return fmt.Sprintf(notificationMessageTemplate, ruleName, count, dataSource, tenantID, windowHours)
}
