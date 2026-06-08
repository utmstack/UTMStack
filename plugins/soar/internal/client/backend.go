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
)

const internalKeyHeader = "X-Internal-Key"

type BackendClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func New() *BackendClient {
	cfg := plugins.PluginCfg("com.utmstack")
	raw := cfg.Get("backend").String()
	if !strings.HasPrefix(raw, "http://") && !strings.HasPrefix(raw, "https://") {
		raw = "http://" + raw
	}
	return &BackendClient{
		baseURL:     raw,
		internalKey: cfg.Get("internalKey").String(),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
	}
}

// Rule mirrors the backend's RuleResponse shape but keeps `conditions` as raw
// JSON — the plugin's eval layer parses it on its own, so re-encoding here would
// be wasted work. Identity is RelPath (the flow file path); there is no integer
// ID. Commands is per-flow; correlate emits one execution per command.
type Rule struct {
	RelPath        string          `json:"relPath"`
	Name           string          `json:"name"`
	Conditions     json.RawMessage `json:"conditions"`
	Commands       []string        `json:"commands"`
	Active         bool            `json:"active"`
	AgentPlatform  string          `json:"agentPlatform"`
	DefaultAgent   string          `json:"defaultAgent"`
	Shell          string          `json:"shell"`
	ExcludedAgents []string        `json:"excludedAgents"`
}

type Execution struct {
	ID               int64  `json:"id"`
	RulePath         string `json:"rulePath"`
	AlertID          string `json:"alertId"`
	Command          string `json:"command"`
	Agent            string `json:"agent"`
	ExecutionStatus  string `json:"executionStatus"`
	ExecutionRetries int    `json:"executionRetries"`
}

// Match server-side enum names from backend/modules/soar/domain/execution.go.
const (
	StatusPending  = "PENDING"
	StatusExecuted = "EXECUTED"
	StatusFailed   = "FAILED"

	CauseAgentOffline  = "AGENT_OFFLINE"
	CauseAgentNotFound = "AGENT_NOT_FOUND"
	CauseUnknown       = "UNKNOWN"
)

type CreateExecutionRequest struct {
	RulePath string `json:"rulePath"`
	AlertID  string `json:"alertId"`
	Command  string `json:"command"`
	Agent    string `json:"agent"`
}

type UpdateExecutionRequest struct {
	ExecutionStatus   *string `json:"executionStatus,omitempty"`
	CommandResult     *string `json:"commandResult,omitempty"`
	NonExecutionCause *string `json:"nonExecutionCause,omitempty"`
	IncrementRetries  bool    `json:"incrementRetries,omitempty"`
}

func (c *BackendClient) ListActiveRules(ctx context.Context) ([]Rule, error) {
	body, err := c.do(ctx, http.MethodGet, "/api/soar/rules?active.equals=true&size=1000", nil)
	if err != nil {
		return nil, err
	}
	var rules []Rule
	if err := json.Unmarshal(body, &rules); err != nil {
		return nil, catcher.Error("failed to decode rules response", err, nil)
	}
	return rules, nil
}

func (c *BackendClient) ListAgentsByPlatform(ctx context.Context, platform string) ([]string, error) {
	if platform == "" {
		return nil, nil
	}
	url := fmt.Sprintf("/api/soar/agents?platform=%s", platform)
	body, err := c.do(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var names []string
	if err := json.Unmarshal(body, &names); err != nil {
		return nil, catcher.Error("failed to decode agents response", err, map[string]any{"platform": platform})
	}
	return names, nil
}

func (c *BackendClient) ListPendingExecutions(ctx context.Context) ([]Execution, error) {
	body, err := c.do(ctx, http.MethodGet, "/api/soar/rule-executions?executionStatus.equals=PENDING&size=1000", nil)
	if err != nil {
		return nil, err
	}
	var execs []Execution
	if err := json.Unmarshal(body, &execs); err != nil {
		return nil, catcher.Error("failed to decode executions response", err, nil)
	}
	return execs, nil
}

func (c *BackendClient) CreateExecution(ctx context.Context, req CreateExecutionRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return catcher.Error("failed to marshal create-execution payload", err, nil)
	}
	_, err = c.do(ctx, http.MethodPost, "/api/soar/rule-executions", payload)
	return err
}

func (c *BackendClient) UpdateExecution(ctx context.Context, id int64, req UpdateExecutionRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return catcher.Error("failed to marshal update-execution payload", err, nil)
	}
	url := fmt.Sprintf("/api/soar/rule-executions/%d", id)
	_, err = c.do(ctx, http.MethodPatch, url, payload)
	return err
}

func (c *BackendClient) do(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, catcher.Error("failed to create backend request", err, map[string]any{"path": path})
	}
	req.Header.Set(internalKeyHeader, c.internalKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, catcher.Error("backend request failed", err, map[string]any{"path": path})
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, catcher.Error("failed to read backend response", err, map[string]any{"path": path})
	}

	if resp.StatusCode >= 400 {
		return nil, catcher.Error("backend returned error status", nil, map[string]any{
			"path":   path,
			"status": resp.StatusCode,
			"body":   string(respBody),
		})
	}
	return respBody, nil
}
