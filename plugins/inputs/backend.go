package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const (
	apiKeyAuthPath = "/api/v1/api-keys/authenticate"
	apiKeyHeader   = "Utm-Api-Key" // header / gRPC metadata key a direct pusher presents
	internalHeader = "X-Internal-Key"
	apiKeyCacheTTL = 60 * time.Second // how long a validated key is trusted before re-checking
)

type apiKeyValidator struct {
	client *http.Client
	mu     sync.RWMutex
	seen   map[string]time.Time // apiKey -> last successful validation
}

var apiKeys = &apiKeyValidator{
	client: &http.Client{Timeout: 15 * time.Second},
	seen:   make(map[string]time.Time),
}

// valid reports whether apiKey is currently a valid backend API key for clientIP.
func (v *apiKeyValidator) valid(apiKey, clientIP string) bool {
	if apiKey == "" {
		return false
	}
	v.mu.RLock()
	at, ok := v.seen[apiKey]
	v.mu.RUnlock()
	if ok && time.Since(at) < apiKeyCacheTTL {
		return true
	}
	if !v.authenticate(apiKey, clientIP) {
		return false
	}
	v.mu.Lock()
	v.seen[apiKey] = time.Now()
	v.mu.Unlock()
	return true
}

func (v *apiKeyValidator) authenticate(apiKey, clientIP string) bool {
	cfg := plugins.PluginCfg("com.utmstack")
	backend := strings.TrimRight(cfg.Get("backend").String(), "/")
	if backend != "" && !strings.HasPrefix(backend, "http://") && !strings.HasPrefix(backend, "https://") {
		backend = "http://" + backend
	}
	internalKey := cfg.Get("internalKey").String()
	if backend == "" || internalKey == "" {
		_ = catcher.Error("api key auth unavailable: backend or internalKey not configured", nil, map[string]any{"process": "plugin_com.utmstack.inputs"})
		return false
	}

	body, err := json.Marshal(map[string]string{"apiKey": apiKey, "clientIp": clientIP})
	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, backend+apiKeyAuthPath, bytes.NewReader(body))
	if err != nil {
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalHeader, internalKey)

	resp, err := v.client.Do(req)
	if err != nil {
		_ = catcher.Error("api key validation request failed", err, map[string]any{"process": "plugin_com.utmstack.inputs"})
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	return resp.StatusCode == http.StatusOK
}
