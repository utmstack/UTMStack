package infrastructure

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/utmstack/utmstack/backend/pkg/http/middleware"
)

type PlaygroundClient struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func NewPlaygroundClient(baseURL, internalKey string) *PlaygroundClient {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return &PlaygroundClient{
		baseURL:     baseURL,
		internalKey: internalKey,
		httpClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *PlaygroundClient) do(ctx context.Context, method, path string, body []byte) (int, []byte, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return 0, nil, fmt.Errorf("playground client: build request %s %s: %w", method, path, err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(middleware.InternalKeyHeader, c.internalKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, nil, fmt.Errorf("playground client: http do %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("playground client: read body %s %s: %w", method, path, err)
	}
	return resp.StatusCode, respBody, nil
}

func (c *PlaygroundClient) DeleteFilters(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodDelete, "/playground/filters", nil)
}

func (c *PlaygroundClient) DeleteRules(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodDelete, "/playground/rules", nil)
}

func (c *PlaygroundClient) CopyFilters(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/playground/filters/copy", nil)
}

func (c *PlaygroundClient) CopyRules(ctx context.Context) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/playground/rules/copy", nil)
}

func (c *PlaygroundClient) WriteFilter(ctx context.Context, name, content string) (int, []byte, error) {
	body, err := json.Marshal(map[string]string{"name": name, "content": content})
	if err != nil {
		return 0, nil, fmt.Errorf("playground client: marshal write-filter body: %w", err)
	}
	return c.do(ctx, http.MethodPost, "/playground/filters", body)
}

func (c *PlaygroundClient) WriteRule(ctx context.Context, name, content string) (int, []byte, error) {
	body, err := json.Marshal(map[string]string{"name": name, "content": content})
	if err != nil {
		return 0, nil, fmt.Errorf("playground client: marshal write-rule body: %w", err)
	}
	return c.do(ctx, http.MethodPost, "/playground/rules", body)
}

func (c *PlaygroundClient) Run(ctx context.Context, log json.RawMessage) (int, []byte, error) {
	return c.do(ctx, http.MethodPost, "/playground/run", []byte(log))
}
