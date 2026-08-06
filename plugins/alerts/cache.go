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
	enrichmentPath      = "/api/v1/datasources/enrichment"
)

type RuleSnapshot struct {
	ID         uint64
	Name       string
	Conditions []FilterType
	TagNames   []string
}

type activeRuleWire struct {
	TenantID        string          `json:"tenantId"`
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
	rules map[string][]RuleSnapshot
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

// Snapshot returns one tenant's rules. This process holds every tenant's, and a
// rule that fired on another tenant's alert would be that tenant's tags leaking
// into someone else's data.
func (c *ruleCache) Snapshot(tenant string) []RuleSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rules := c.rules[tenant]
	if len(rules) == 0 {
		return nil
	}
	out := make([]RuleSnapshot, len(rules))
	copy(out, rules)
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

	next := make(map[string][]RuleSnapshot, len(wire))
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
		next[r.TenantID] = append(next[r.TenantID], RuleSnapshot{
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

type datasourceEnrichment struct {
	GroupID   *uint64
	GroupName string
	Labels    []string
}

type enrichmentWire struct {
	TenantID  string   `json:"tenantId"`
	Name      string   `json:"name"`
	DataType  string   `json:"dataType"`
	GroupID   *uint64  `json:"groupId"`
	GroupName string   `json:"groupName"`
	Labels    []string `json:"labels"`
}

type datasourceCache struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
	refresh     time.Duration

	mu     sync.RWMutex
	byName map[dsKey]datasourceEnrichment
}

// dsKey is what the backend's uniqueness actually is: a datasource name is
// unique inside a tenant, and two tenants running a host of the same name are
// two different assets.
type dsKey struct{ tenant, name string }

func newDatasourceCache() *datasourceCache {
	cfg := plugins.PluginCfg("com.utmstack")
	base := cfg.Get("backend").String()
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	refresh := time.Duration(defaultRefreshSec) * time.Second
	if v := cfg.Get("rulesRefreshSec").Int(); v > 0 {
		refresh = time.Duration(v) * time.Second
	}
	return &datasourceCache{
		baseURL:     base,
		internalKey: cfg.Get("internalKey").String(),
		httpClient:  &http.Client{Timeout: rulesRequestTimeout},
		refresh:     refresh,
	}
}

func (c *datasourceCache) Lookup(tenant, name string) (datasourceEnrichment, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.byName[dsKey{tenant, name}]
	return e, ok
}

func (c *datasourceCache) Refresh(ctx context.Context) error {
	if c.baseURL == "" || c.internalKey == "" {
		return catcher.Error("datasource cache: backend URL or internal key missing", nil, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+enrichmentPath, nil)
	if err != nil {
		return catcher.Error("datasource cache: build request failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	req.Header.Set(internalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return catcher.Error("datasource cache: request failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return catcher.Error("datasource cache: read body failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}
	if resp.StatusCode >= 400 {
		return catcher.Error("datasource cache: backend returned error", nil, map[string]any{
			"status":  resp.StatusCode,
			"body":    string(body),
			"process": "plugin_com.utmstack.alerts",
		})
	}

	var wire []enrichmentWire
	if err := json.Unmarshal(body, &wire); err != nil {
		return catcher.Error("datasource cache: decode failed", err, map[string]any{"process": "plugin_com.utmstack.alerts"})
	}

	next := make(map[dsKey]datasourceEnrichment, len(wire))
	for _, w := range wire {
		// A host's group is the same across its data types, so a repeated name
		// within a tenant just overwrites with equivalent enrichment.
		next[dsKey{w.TenantID, w.Name}] = datasourceEnrichment{
			GroupID:   w.GroupID,
			GroupName: w.GroupName,
			Labels:    w.Labels,
		}
	}

	c.mu.Lock()
	c.byName = next
	c.mu.Unlock()
	return nil
}

func (c *datasourceCache) Run(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("datasource cache: recovered from panic in Run", nil, map[string]any{
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
