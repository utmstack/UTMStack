package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/soar/domain"
)

// HTTP performs an outgoing HTTP call. As an executor kind, a non-2xx status
// is an error and the response body is ignored (logged into Result). As an
// enrichment kind, the response body must parse as JSON and is returned as the
// node's output.
type HTTP struct {
	client *http.Client
}

func NewHTTP() *HTTP {
	return &HTTP{client: &http.Client{Timeout: 30 * time.Second}}
}

func (HTTP) Type() string { return "http" }

type httpParams struct {
	Method  string            `json:"method,omitempty"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    json.RawMessage   `json:"body,omitempty"`
	Timeout int               `json:"timeoutSec,omitempty"`
}

func (h *HTTP) Execute(ctx context.Context, exec *domain.SoarExecution) (json.RawMessage, error) {
	var p httpParams
	if len(exec.Params) > 0 {
		if err := json.Unmarshal(exec.Params, &p); err != nil {
			return nil, fmt.Errorf("soar http: params: %w", err)
		}
	}
	if strings.TrimSpace(p.URL) == "" {
		return nil, errors.New("soar http: url is required")
	}
	if p.Method == "" {
		if len(p.Body) > 0 {
			p.Method = http.MethodPost
		} else {
			p.Method = http.MethodGet
		}
	}

	client := h.client
	if p.Timeout > 0 {
		client = &http.Client{Timeout: time.Duration(p.Timeout) * time.Second}
	}

	var body io.Reader
	if len(p.Body) > 0 {
		body = bytes.NewReader(p.Body)
	}
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(p.Method), p.URL, body)
	if err != nil {
		return nil, fmt.Errorf("soar http: build request: %w", err)
	}
	if len(p.Body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range p.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("soar http: %s %s: %w", p.Method, p.URL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("soar http: read body: %w", err)
	}
	exec.Result = fmt.Sprintf("%d %s\n%s", resp.StatusCode, resp.Status, truncate(string(raw), 4096))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("soar http: %s %s returned %d", p.Method, p.URL, resp.StatusCode)
	}

	if exec.Kind != domain.NodeKindEnrichment {
		return nil, nil
	}
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return json.RawMessage(`{}`), nil
	}
	var probe any
	if err := json.Unmarshal(trimmed, &probe); err != nil {
		return nil, fmt.Errorf("soar http enrichment: body is not JSON: %w", err)
	}
	return json.RawMessage(trimmed), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
