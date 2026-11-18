// edr/internal/feeds/creds.go
package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Backend appconfig keys where the platform stores its ThreatWinds
// credentials (registered and written by the backend / plugins-feeds flow —
// this daemon only reads them).
const (
	twAPIKeyKey    = "utmstack.tw.apiKey"
	twAPISecretKey = "utmstack.tw.apiSecret"
)

// FetchCredentials reads the ThreatWinds api-key/api-secret from the backend
// appconfig API, authenticated with the stack INTERNAL_KEY.
func FetchCredentials(ctx context.Context, utmHost, internalKey string, httpc *http.Client) (string, string, error) {
	get := func(key string) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/config/%s", utmHost, key), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("X-Internal-Key", internalKey)
		resp, err := httpc.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("config %s: status %d", key, resp.StatusCode)
		}
		var out struct {
			Value string `json:"value"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return "", err
		}
		return out.Value, nil
	}
	k, err := get(twAPIKeyKey)
	if err != nil {
		return "", "", err
	}
	s, err := get(twAPISecretKey)
	if err != nil {
		return "", "", err
	}
	return k, s, nil
}
