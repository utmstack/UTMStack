// Package eventprocessor talks to the external event-processor-manager service.
// The legacy Spring backend pushed module config changes and validation requests
// over HTTP with a shared `internal-key` header; this client preserves that
// contract so the Go port keeps wire compatibility.
package eventprocessor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// InternalKeyHeader matches Constants.EVENT_PROCESSOR_INTERNAL_KEY_HEADER in
// the legacy backend.
const InternalKeyHeader = "internal-key"

type Client struct {
	baseURL     string
	internalKey string
	http        *http.Client
}

func NewClient(host string, port int, internalKey string) *Client {
	return &Client{
		baseURL:     fmt.Sprintf("http://%s:%d", host, port),
		internalKey: internalKey,
		http:        &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) post(ctx context.Context, path string, query url.Values, body any) (int, []byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return 0, nil, fmt.Errorf("marshal: %w", err)
	}

	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(buf))
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")
	if c.internalKey != "" {
		req.Header.Set(InternalKeyHeader, c.internalKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 300 {
		return resp.StatusCode, respBody, fmt.Errorf("event-processor %s %d: %s", path, resp.StatusCode, string(respBody))
	}
	return resp.StatusCode, respBody, nil
}
