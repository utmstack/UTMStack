package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/agent-manager/config"
	"github.com/utmstack/UTMStack/agent-manager/models"
)

const (
	defaultDatasourcePingEvery = 60 * time.Second
	datasourcePingPath         = "/api/v1/datasources/ping"
	internalKeyHeader          = "X-Internal-Key"
)

type dsPingEntry struct {
	TenantID   string          `json:"tenantId,omitempty"`
	Name       string          `json:"name"`
	DataType   string          `json:"dataType,omitempty"`
	SourceKind string          `json:"sourceKind"`
	IP         string          `json:"ip,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	LastPingAt *time.Time      `json:"lastPingAt,omitempty"`
}

type dsPingRequest struct {
	Datasources []dsPingEntry `json:"datasources"`
}

var datasourceHTTPClient = &http.Client{Timeout: 30 * time.Second}

func (s *LastSeenService) StartDatasourceSync() {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("recovered from panic in datasource sync", nil, map[string]any{"panic": r, "process": "agent-manager"})
		}
	}()

	if config.UTMHost == "" || config.InternalKey == "" {
		catcher.Info("datasource sync disabled: UTM_HOST or INTERNAL_KEY not set", map[string]any{"process": "agent-manager"})
		return
	}

	ticker := time.NewTicker(datasourcePingEvery())
	defer ticker.Stop()
	for range ticker.C {
		s.syncDatasources()
	}
}

func (s *LastSeenService) syncDatasources() {
	entries := s.buildDatasourceEntries()
	if len(entries) == 0 {
		return
	}

	body, err := json.Marshal(dsPingRequest{Datasources: entries})
	if err != nil {
		_ = catcher.Error("failed to marshal datasource ping", err, map[string]any{"process": "agent-manager"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.UTMHost+datasourcePingPath, bytes.NewReader(body))
	if err != nil {
		_ = catcher.Error("failed to build datasource ping request", err, map[string]any{"process": "agent-manager"})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(internalKeyHeader, config.InternalKey)

	resp, err := datasourceHTTPClient.Do(req)
	if err != nil {
		_ = catcher.Error("datasource ping request failed", err, map[string]any{"process": "agent-manager"})
		return
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_ = catcher.Error("datasource ping returned error status", fmt.Errorf("status %d", resp.StatusCode), map[string]any{"process": "agent-manager"})
	}
}

func (s *LastSeenService) buildDatasourceEntries() []dsPingEntry {
	s.CacheAgentLastSeenMutex.Lock()
	agentSeen := make(map[uint]time.Time, len(s.CacheAgentLastSeen))
	for id, ls := range s.CacheAgentLastSeen {
		agentSeen[id] = ls.LastPing
	}
	s.CacheAgentLastSeenMutex.Unlock()

	return s.agentEntries(agentSeen)
}

func (s *LastSeenService) agentEntries(seen map[uint]time.Time) []dsPingEntry {
	if len(seen) == 0 {
		return nil
	}
	ids := make([]uint, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}

	var agents []models.Agent
	if _, err := s.DBConnection.GetAll(&agents, "id IN ? AND deleted_at IS NULL", ids); err != nil {
		_ = catcher.Error("failed to load agents for datasource ping", err, map[string]any{"process": "agent-manager"})
		return nil
	}

	out := make([]dsPingEntry, 0, len(agents))
	for _, a := range agents {
		if a.Hostname == "" {
			continue
		}
		ts := seen[a.ID]
		out = append(out, dsPingEntry{
			TenantID:   tenantOrDefault(a.TenantID),
			Name:       a.Hostname,
			DataType:   agentDataType(a.Os),
			SourceKind: "agent",
			IP:         a.Ip,
			Metadata: jsonMetadata(map[string]string{
				"agentId":         strconv.FormatUint(uint64(a.ID), 10),
				"osPlatform":      a.Platform,
				"os":              a.Os,
				"mac":             a.Mac,
				"addresses":       a.Addresses,
				"version":         a.Version,
				"aliases":         a.Aliases,
				"noRemoteControl": strconv.FormatBool(a.NoRemoteControl),
			}),
			LastPingAt: &ts,
		})
	}
	return out
}

func agentDataType(osType string) string {
	switch strings.ToLower(osType) {
	case "windows":
		return "wineventlog"
	case "macos", "darwin":
		return "macos"
	default:
		return "linux"
	}
}

func jsonMetadata(m map[string]string) json.RawMessage {
	out := make(map[string]string, len(m))
	for k, v := range m {
		if v != "" {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	b, err := json.Marshal(out)
	if err != nil {
		return nil
	}
	return b
}

func datasourcePingEvery() time.Duration {
	if v := os.Getenv("DATASOURCE_PING_INTERVAL_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return defaultDatasourcePingEvery
}
