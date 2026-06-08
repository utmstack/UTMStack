package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const (
	internalKeyHeader   = "X-Internal-Key"
	defaultRefreshSec   = 60
	rulesRequestTimeout = 10 * time.Second
	activeRulesPath     = "/api/v1/internal/alert-tag-rules/active"
)

// RuleSnapshot is the per-rule view the plugin needs at evaluation time.
// Conditions is left as raw JSON until used so we don't decode work we won't
// run (a rule with empty conditions never reaches the evaluator).
type RuleSnapshot struct {
	ID         uint64
	Name       string
	Conditions []FilterType
	TagNames   []string
}

// activeRuleWire matches the backend dto.ActiveAlertTagRule payload.
type activeRuleWire struct {
	ID              uint64          `json:"id"`
	Name            string          `json:"name"`
	Conditions      json.RawMessage `json:"conditions"`
	AppliedTagNames []string        `json:"appliedTagNames"`
}

type ruleCache struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
	refresh     time.Duration

	mu    sync.RWMutex
	rules []RuleSnapshot
}

func newRuleCache() *ruleCache {
	cfg := plugins.PluginCfg("com.utmstack")
	base := cfg.Get("backend").String()
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	refresh := time.Duration(defaultRefreshSec) * time.Second
	if v := cfg.Get("rulesRefreshSec").Int(); v > 0 {
		refresh = time.Duration(v) * time.Second
	}
	return &ruleCache{
		baseURL:     base,
		internalKey: cfg.Get("internalKey").String(),
		httpClient:  &http.Client{Timeout: rulesRequestTimeout},
		refresh:     refresh,
	}
}

func (c *ruleCache) Snapshot() []RuleSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.rules) == 0 {
		return nil
	}
	out := make([]RuleSnapshot, len(c.rules))
	copy(out, c.rules)
	return out
}

func (c *ruleCache) Refresh(ctx context.Context) error {
	if c.baseURL == "" || c.internalKey == "" {
		return catcher.Error("rule cache: backend URL or internal key missing", nil, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+activeRulesPath, nil)
	if err != nil {
		return catcher.Error("rule cache: build request failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	req.Header.Set(internalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("rule cache: request failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return catcher.Error("rule cache: read body failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	if resp.StatusCode >= 400 {
		return catcher.Error("rule cache: backend returned error", nil, map[string]any{
			"status":  resp.StatusCode,
			"body":    string(body),
			"process": "plugin_com.utmstack.alerts",
		})
	}

	var wire []activeRuleWire
	if err := json.Unmarshal(body, &wire); err != nil {
		return catcher.Error("rule cache: decode failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}

	next := make([]RuleSnapshot, 0, len(wire))
	for _, r := range wire {
		var conditions []FilterType
		if len(r.Conditions) > 0 {
			if err := json.Unmarshal(r.Conditions, &conditions); err != nil {
				_ = catcher.Error("rule cache: skipping rule with invalid conditions JSON", err, map[string]any{
					"rule":    r.Name,
					"process": "plugin_com.utmstack.alerts",
				})
				continue
			}
		}
		next = append(next, RuleSnapshot{
			ID:         r.ID,
			Name:       r.Name,
			Conditions: conditions,
			TagNames:   r.AppliedTagNames,
		})
	}

	c.mu.Lock()
	c.rules = next
	c.mu.Unlock()
	return nil
}

func (c *ruleCache) Run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("rule cache: recovered from panic in Run", nil, map[string]any{
				"panic":   r,
				"process": "plugin_com.utmstack.alerts",
			})
		}
	}()

	ticker := time.NewTicker(c.refresh)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			refreshCtx, cancel := context.WithTimeout(context.Background(), rulesRequestTimeout)
			_ = c.Refresh(refreshCtx)
			cancel()
		}
	}
}
