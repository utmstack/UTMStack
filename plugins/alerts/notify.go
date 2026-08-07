package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const (
	notifyPath    = "/api/v1/internal/alerts/notify"
	notifyTimeout = 5 * time.Second
)

type notifier struct {
	baseURL     string
	internalKey string
	httpClient  *http.Client
}

func newNotifier() *notifier {
	cfg := plugins.PluginCfg("com.utmstack")
	base := cfg.Get("backend").String()
	if base != "" && !strings.HasPrefix(base, "http://") && !strings.HasPrefix(base, "https://") {
		base = "http://" + base
	}
	return &notifier{
		baseURL:     base,
		internalKey: cfg.Get("internalKey").String(),
		httpClient:  &http.Client{Timeout: notifyTimeout},
	}
}

func (n *notifier) Notify(tenantID, alertID, parentID string) {
	if n == nil || n.baseURL == "" || n.internalKey == "" || alertID == "" {
		return
	}
	if parentID != "" {
		return
	}

	body, err := json.Marshal(map[string]string{"alertId": alertID})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+notifyPath, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set(internalKeyHeader, n.internalKey)
	req.Header.Set(tenantHeader, tenantID)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		_ = catcher.Error("alert notification request failed", err, map[string]any{
			"alert":   alertID,
			"process": processName,
		})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		_ = catcher.Error("backend refused the alert notification", nil, map[string]any{
			"status":  resp.StatusCode,
			"alert":   alertID,
			"process": processName,
		})
	}
}
