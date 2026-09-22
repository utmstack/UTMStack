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
		if neverExposeName(t.Name) {
			continue
		}
		schema, _ := t.InputSchema.(map[string]any)
		specs = append(specs, ToolSpec{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
			ReadOnly:    isReadOnly(t),
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

// isReadOnly trusts only the backend's explicit ReadOnlyHint (fail-closed).
// The old name-based verb guess let tenants.terminate through as "read-only"
// since "terminate" wasn't in its whitelist.
func isReadOnly(t *mcp.Tool) bool {
	return t.Annotations != nil && t.Annotations.ReadOnlyHint
}

// neverExposeName hard-excludes prefixes with no capability group of their
// own (tenants.*, platform.*) as a second layer beyond isReadOnly.
func neverExposeName(name string) bool {
	prefix := name
	if i := strings.IndexByte(name, '.'); i >= 0 {
		prefix = name[:i]
	}
	return prefix == "tenants" || prefix == "platform"
}
