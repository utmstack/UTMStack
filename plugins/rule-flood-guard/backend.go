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

const notificationMessageTemplate = "Correlation rule '%s' generated over %d un-deduplicated alerts and was automatically disabled to prevent alert flooding. Mark false positives via Alert Tag Rules, or add deduplicateBy/groupBy to the rule, then re-enable it."

const maxNotificationMessageLen = 280

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

type ruleFilterResult struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	RuleActive bool   `json:"ruleActive"`
}

func (c *backendClient) resolveRule(ctx context.Context, ruleName string) (*ruleFilterResult, error) {
	endpoint := c.baseURL + "/api/correlation-rule/search-by-filters?name=" + url.QueryEscape(ruleName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, catcher.Error("search-by-filters call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "ruleName": ruleName,
		})
	}

	var candidates []ruleFilterResult
	if err := json.NewDecoder(resp.Body).Decode(&candidates); err != nil {
		return nil, err
	}

	var exact *ruleFilterResult
	matches := 0
	for i := range candidates {
		if strings.EqualFold(candidates[i].Name, ruleName) {
			matches++
			exact = &candidates[i]
		}
	}
	if matches != 1 {
		if matches > 1 {
			_ = catcher.Error("rule-flood-guard: ambiguous rule name collision, skipping disable", nil, map[string]any{
				"ruleName": ruleName, "matches": matches,
			})
		}
		return nil, nil
	}
	return exact, nil
}

func (c *backendClient) Deactivate(ctx context.Context, ruleName string) (bool, error) {
	rule, err := c.resolveRule(ctx, ruleName)
	if err != nil {
		return false, err
	}
	if rule == nil {
		// Not found, or ambiguous — already logged by resolveRule.
		return false, nil
	}
	if !rule.RuleActive {
		// Already disabled (manually, or by a previous/overlapping cycle).
		return false, nil
	}

	endpoint := fmt.Sprintf("%s/api/correlation-rule/activate-deactivate?id=%d&active=false", c.baseURL, rule.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Utm-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return false, catcher.Error("activate-deactivate call returned error status", nil, map[string]any{
			"status": resp.StatusCode, "body": string(body), "ruleName": ruleName, "id": rule.Id,
		})
	}
	return true, nil
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
	endpoint := c.baseURL + "/api/notifications"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Utm-Internal-Key", c.internalKey)
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
	msg := fmt.Sprintf(notificationMessageTemplate, ruleName, threshold)
	if len(msg) <= maxNotificationMessageLen {
		return msg
	}

	overflow := len(msg) - maxNotificationMessageLen
	cut := len(ruleName) - overflow - 3
	if cut < 1 {
		cut = 1
	}
	name := ruleName[:cut] + "..."
	return fmt.Sprintf(notificationMessageTemplate, name, threshold)
}
