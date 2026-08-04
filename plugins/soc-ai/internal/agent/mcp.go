package agent

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ToolBroker struct {
	endpoint string
	key      string
	tenantID string

	mu      sync.Mutex
	session *mcp.ClientSession
}

func NewToolBroker(backendURL, internalKey, tenantID string) *ToolBroker {
	endpoint := strings.TrimRight(backendURL, "/") + "/api/v1/mcp"
	return &ToolBroker{endpoint: endpoint, key: internalKey, tenantID: tenantID}
}

type keyRoundTripper struct {
	base     http.RoundTripper
	key      string
	tenantID string
}

func (k keyRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("X-Internal-Key", k.key)
	if k.tenantID != "" {
		r.Header.Set("X-Tenant-Id", k.tenantID)
	}
	return k.base.RoundTrip(r)
}

func (b *ToolBroker) ensure(ctx context.Context) (*mcp.ClientSession, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.session != nil {
		return b.session, nil
	}
	httpClient := &http.Client{
		Timeout: 90 * time.Second,
		Transport: keyRoundTripper{
			base: &http.Transport{
				MaxIdleConns:        20,
				IdleConnTimeout:     90 * time.Second,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // backend may use a self-signed internal cert
				MaxIdleConnsPerHost: 10,
			},
			key:      b.key,
			tenantID: b.tenantID,
		},
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "com.utmstack.soc-ai", Version: "1.0.0"}, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:             b.endpoint,
		HTTPClient:           httpClient,
		DisableStandaloneSSE: true, // request/response only; we don't need server-initiated messages
	}
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		return nil, err
	}
	b.session = session
	return session, nil
}

func (b *ToolBroker) reset() {
	b.mu.Lock()
	if b.session != nil {
		_ = b.session.Close()
		b.session = nil
	}
	b.mu.Unlock()
}

func (b *ToolBroker) Close() { b.reset() }

func (b *ToolBroker) ListSpecs(ctx context.Context) ([]ToolSpec, error) {
	res, err := b.listOnce(ctx)
	if err != nil {
		// One reconnect attempt — the session may have expired.
		b.reset()
		res, err = b.listOnce(ctx)
		if err != nil {
			return nil, err
		}
	}
	specs := make([]ToolSpec, 0, len(res.Tools))
	for _, t := range res.Tools {
		schema, _ := t.InputSchema.(map[string]any)
		readOnly := !mutatingName(t.Name)
		if t.Annotations != nil && t.Annotations.ReadOnlyHint {
			readOnly = true
		}
		specs = append(specs, ToolSpec{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
			ReadOnly:    readOnly,
		})
	}
	return specs, nil
}

func (b *ToolBroker) listOnce(ctx context.Context) (*mcp.ListToolsResult, error) {
	session, err := b.ensure(ctx)
	if err != nil {
		return nil, err
	}
	return session.ListTools(ctx, nil)
}

func (b *ToolBroker) Call(ctx context.Context, name string, args json.RawMessage) (string, bool, error) {
	res, err := b.callOnce(ctx, name, args)
	if err != nil {
		b.reset()
		res, err = b.callOnce(ctx, name, args)
		if err != nil {
			return "", true, err
		}
	}
	var sb strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			sb.WriteString(tc.Text)
		}
	}
	out := sb.String()
	if out == "" && res.StructuredContent != nil {
		if raw, mErr := json.Marshal(res.StructuredContent); mErr == nil {
			out = string(raw)
		}
	}
	return out, res.IsError, nil
}

func (b *ToolBroker) callOnce(ctx context.Context, name string, args json.RawMessage) (*mcp.CallToolResult, error) {
	session, err := b.ensure(ctx)
	if err != nil {
		return nil, err
	}
	var arguments any
	if len(args) > 0 {
		arguments = json.RawMessage(args)
	}
	return session.CallTool(ctx, &mcp.CallToolParams{Name: name, Arguments: arguments})
}

func mutatingName(name string) bool {
	tokens := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return r == '.' || r == '_' || r == '-' || r == '/'
	})
	for _, tok := range tokens {
		if mutatingVerbs[tok] {
			return true
		}
	}
	return false
}

var mutatingVerbs = map[string]bool{
	"create": true, "update": true, "delete": true, "convert": true,
	"set": true, "add": true, "remove": true, "disable": true, "enable": true,
	"execute": true, "run": true, "shutdown": true, "isolate": true,
	"block": true, "save": true, "generate": true, "evaluate": true,
	"activate": true, "deactivate": true, "job": true, "jobs": true,
	"assign": true, "change": true, "send": true, "mark": true,
}
