package auth

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"

	"github.com/utmstack/UTMStack/log-input/agent"
	"github.com/utmstack/UTMStack/log-input/config"
)

const (
	maxMessageSize = 20 * 1024 * 1024
	listPageSize   = 100000

	upstreamTimeout = 15 * time.Second

	apiKeyAuthPath = "/api/v1/api-keys/authenticate"
	internalHeader = "X-Internal-Key"
)

type agentManagerClient struct {
	cfg *config.Config
}

func newAgentManagerClient(cfg *config.Config) *agentManagerClient {
	return &agentManagerClient{cfg: cfg}
}

func (c *agentManagerClient) listKeys(ctx context.Context, typ string) (map[uint64]string, error) {
	// Self-signed certificate on an internal address; the internal key is what
	// authenticates the call.
	creds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

	conn, err := grpc.NewClient(c.cfg.AgentManager,
		grpc.WithTransportCredentials(creds),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)),
	)
	if err != nil {
		return nil, fmt.Errorf("dialling the agent-manager: %w", err)
	}
	defer func() { _ = conn.Close() }()

	cCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
	defer cancel()
	cCtx = metadata.AppendToOutgoingContext(cCtx, "internal-key", c.cfg.InternalKey)

	req := &agent.ListRequest{PageNumber: 1, PageSize: listPageSize}

	switch strings.ToLower(typ) {
	case "agent":
		resp, err := agent.NewAgentServiceClient(conn).ListAgents(cCtx, req)
		if err != nil {
			return nil, err
		}
		out := make(map[uint64]string, len(resp.Rows))
		for _, row := range resp.Rows {
			out[uint64(row.Id)] = row.AgentKey
		}
		return out, nil

	case "collector":
		resp, err := agent.NewCollectorServiceClient(conn).ListCollector(cCtx, req)
		if err != nil {
			return nil, err
		}
		out := make(map[uint64]string, len(resp.Rows))
		for _, row := range resp.Rows {
			out[uint64(row.Id)] = row.CollectorKey
		}
		return out, nil
	}

	return nil, fmt.Errorf("unknown connector type %q", typ)
}

type backendClient struct {
	cfg    *config.Config
	client *http.Client
}

func newBackendClient(cfg *config.Config) *backendClient {
	return &backendClient{
		cfg:    cfg,
		client: &http.Client{Timeout: upstreamTimeout},
	}
}

func (c *backendClient) authenticate(ctx context.Context, apiKey, clientIP string) bool {
	base := strings.TrimRight(c.cfg.Backend, "/")
	if base == "" {
		return false
	}
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}

	body, err := json.Marshal(map[string]string{"apiKey": apiKey, "clientIp": clientIP})
	if err != nil {
		return false
	}

	cCtx, cancel := context.WithTimeout(ctx, upstreamTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cCtx, http.MethodPost, base+apiKeyAuthPath, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalHeader, c.cfg.InternalKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}
