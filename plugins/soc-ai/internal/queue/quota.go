package queue

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

const (
	quotaPath    = "/api/v1/soc-ai/quota/consume"
	quotaTimeout = 10 * time.Second
)

var quotaClient = &http.Client{Timeout: quotaTimeout}

func consumeQuota(ctx context.Context, backend, internalKey, tenantID string) bool {
	if backend == "" || internalKey == "" || tenantID == "" {
		return true
	}

	base := strings.TrimRight(backend, "/")
	if !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}

	cCtx, cancel := context.WithTimeout(ctx, quotaTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(cCtx, http.MethodPost, base+quotaPath, nil)
	if err != nil {
		return true
	}
	req.Header.Set("X-Internal-Key", internalKey)
	req.Header.Set("X-Tenant-Id", tenantID)

	resp, err := quotaClient.Do(req)
	if err != nil {
		_ = catcher.Error("cannot reach the AI quota; analysing anyway", err, map[string]any{
			"process": "plugin_com.utmstack.soc-ai",
			"tenant":  tenantID,
		})
		return true
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests {
		return false
	}
	return true
}
