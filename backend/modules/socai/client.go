package socai

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"
)

type SocAIClient struct {
	baseURL      string
	internalKey  string
	httpClient   *http.Client
	streamClient *http.Client
}

func NewSocAIClient(baseURL, internalKey string) *SocAIClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // intentional — mirrors Java parity for internal cluster
	}
	return &SocAIClient{
		baseURL:     baseURL,
		internalKey: internalKey,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		streamClient: &http.Client{
			Transport: transport,
		},
	}
}

func (c *SocAIClient) StreamAgentTask(ctx context.Context, body []byte) (*http.Response, error) {
	url := c.baseURL + "/api/v1/agent/task"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("socai: build stream request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-Internal-Key", c.internalKey)

	resp, err := c.streamClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("socai: stream http do: %w", err)
	}
	return resp, nil
}

func (c *SocAIClient) Analyze(ctx context.Context, alertJSON []byte) (int, []byte, error) {
	url := c.baseURL + "/api/v1/analyze"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(alertJSON))
	if err != nil {
		return 0, nil, fmt.Errorf("socai: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("socai: http do: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("socai: read body: %w", err)
	}

	return resp.StatusCode, body, nil
}
