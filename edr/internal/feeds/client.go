// edr/internal/feeds/client.go
package feeds

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Client downloads indicator lists from the ThreatWinds Feeds API.
type Client struct {
	BaseURL   string
	APIKey    string
	APISecret string
	HTTP      *http.Client
}

func (c *Client) FetchAccumulative(ctx context.Context, level, name string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/feeds/v1/download/list/%s/accumulative/%s", c.BaseURL, level, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("api-key", c.APIKey)
	req.Header.Set("api-secret", c.APISecret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
