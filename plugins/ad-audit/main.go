// Command ad-audit is an UTMStack analysis plugin that keeps an Active Directory
// user inventory. The event processor streams every processed event to it; for
// the relevant Windows security events it updates an in-memory cache and
// periodically flushes the users that changed to the backend's internal ingest
// endpoint. Replaces the legacy `user-auditor` Java microservice.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const pluginName = "com.utmstack.ad-audit"

func main() {
	if plugins.GetCfg("plugin_"+pluginName).Env.Mode != "manager" {
		return
	}

	loadBackendConfig()
	seedCacheFromBackend()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	go flushLoop(ctx, 60*time.Second)

	if err := plugins.InitAnalysisPlugin(pluginName, analyze); err != nil {
		_ = catcher.Error("failed to start analysis plugin", err, map[string]any{"process": "plugin_" + pluginName})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}

const (
	evtUserCreated = "4720" // A user account was created
	evtUserDeleted = "4726" // A user account was deleted
	evtLogon       = "4624" // An account was successfully logged on

	wineventlogDataType = "wineventlog"

	fEventCode  = "eventCode"
	fTargetUser = "eventDataTargetUserName"
	fTargetSID  = "eventDataTargetSid"
	fTargetSID2 = "eventDataTargetUserSid"
	fTargetDom  = "eventDataTargetDomainName"
)

func analyze(event *plugins.Event, _ plugins.Analysis_AnalyzeServer) error {
	if event.GetDataType() != wineventlogDataType {
		return io.EOF
	}
	code := logStr(event, fEventCode)
	switch code {
	case evtUserCreated, evtUserDeleted, evtLogon:
	default:
		return io.EOF
	}

	name := logStr(event, fTargetUser)
	if isSystemAccount(name) {
		return io.EOF
	}
	domain := logStr(event, fTargetDom)
	sid := firstNonEmpty(logStr(event, fTargetSID), logStr(event, fTargetSID2))
	if sid == "" {
		if name == "" {
			return io.EOF
		}
		sid = strings.ToLower(strings.TrimSpace(domain + "\\" + name)) // stable fallback identity
	}

	applyEvent(event.GetTenantId(), sid, name, domain, code, parseTime(event.GetTimestamp()))
	return io.EOF
}

func logStr(event *plugins.Event, key string) string {
	log := event.GetLog()
	if log == nil || log[key] == nil {
		return ""
	}
	return strings.TrimSpace(log[key].GetStringValue())
}

func isSystemAccount(name string) bool {
	switch n := strings.ToUpper(name); {
	case n == "":
		return false
	case strings.HasSuffix(n, "$"):
		return true
	case n == "SYSTEM", n == "LOCAL SERVICE", n == "NETWORK SERVICE", n == "ANONYMOUS LOGON":
		return true
	case strings.HasPrefix(n, "DWM-"), strings.HasPrefix(n, "UMFD-"):
		return true
	}
	return false
}

// ── In-memory cache ───────────────────────────────────────────────────────────

type cachedUser struct {
	TenantID         string
	SID              string
	SamAccountName   string
	Domain           string
	Active           bool
	AccountCreatedAt *time.Time
	LastLogon        *time.Time
	AccountDeletedAt *time.Time
	LastSeen         *time.Time
}

var (
	cacheMu sync.Mutex
	cache   = map[string]*cachedUser{} // "tenant:sid" -> user
	dirty   = map[string]struct{}{}    // ids changed since the last flush
)

func applyEvent(tenantID, sid, name, domain, code string, et *time.Time) {
	id := tenantID + ":" + sid

	cacheMu.Lock()
	defer cacheMu.Unlock()

	cu := cache[id]
	if cu == nil {
		cu = &cachedUser{TenantID: tenantID, SID: sid, Active: true}
		cache[id] = cu
	}
	before := *cu

	if name != "" {
		cu.SamAccountName = name
	}
	if domain != "" {
		cu.Domain = domain
	}
	cu.LastSeen = laterTime(cu.LastSeen, et)
	switch code {
	case evtUserCreated:
		cu.AccountCreatedAt = firstSet(cu.AccountCreatedAt, et)
		cu.Active = true
	case evtLogon:
		cu.LastLogon = laterTime(cu.LastLogon, et)
		cu.Active = true
	case evtUserDeleted:
		cu.Active = false
		cu.AccountDeletedAt = et
	}

	if *cu != before {
		dirty[id] = struct{}{}
	}
}

func parseTime(s string) *time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return &t
		}
	}
	return nil
}

func laterTime(cur, next *time.Time) *time.Time {
	if next != nil && (cur == nil || next.After(*cur)) {
		return next
	}
	return cur
}

func firstSet(cur, next *time.Time) *time.Time {
	if cur != nil {
		return cur
	}
	return next
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ── Backend sync (cache → backend, batched) ───────────────────────────────────

const internalKeyHeader = "X-Internal-Key"

var (
	backendURL  string
	internalKey string
	httpClient  = &http.Client{Timeout: 30 * time.Second}
)

type ingestUser struct {
	TenantID         string     `json:"tenantId"`
	SID              string     `json:"sid"`
	SamAccountName   string     `json:"samAccountName"`
	Domain           string     `json:"domain"`
	Active           *bool      `json:"active"`
	AccountCreatedAt *time.Time `json:"accountCreatedAt"`
	LastLogon        *time.Time `json:"lastLogon"`
	AccountDeletedAt *time.Time `json:"accountDeletedAt"`
	LastSeen         *time.Time `json:"lastSeen"`
}

func (cu *cachedUser) toIngest() ingestUser {
	active := cu.Active
	return ingestUser{
		TenantID: cu.TenantID, SID: cu.SID, SamAccountName: cu.SamAccountName, Domain: cu.Domain,
		Active: &active, AccountCreatedAt: cu.AccountCreatedAt, LastLogon: cu.LastLogon,
		AccountDeletedAt: cu.AccountDeletedAt, LastSeen: cu.LastSeen,
	}
}

func loadBackendConfig() {
	cfg := plugins.PluginCfg("com.utmstack")
	backendURL = strings.TrimRight(cfg.Get("backend").String(), "/")
	internalKey = cfg.Get("internalKey").String()
}

func request(method, path string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, backendURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set(internalKeyHeader, internalKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return httpClient.Do(req)
}

func seedCacheFromBackend() {
	if backendURL == "" {
		return
	}
	resp, err := request(http.MethodGet, "/api/v1/ad-audit/users/sync", nil)
	if err != nil {
		_ = catcher.Error("ad-audit: seed request failed", err, nil)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = catcher.Error("ad-audit: seed returned non-200", fmt.Errorf("status %d", resp.StatusCode), nil)
		return
	}

	var users []ingestUser
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		_ = catcher.Error("ad-audit: decoding seed failed", err, nil)
		return
	}

	cacheMu.Lock()
	for _, u := range users {
		cache[u.TenantID+":"+u.SID] = &cachedUser{
			TenantID: u.TenantID, SID: u.SID, SamAccountName: u.SamAccountName, Domain: u.Domain,
			Active: u.Active == nil || *u.Active, AccountCreatedAt: u.AccountCreatedAt,
			LastLogon: u.LastLogon, AccountDeletedAt: u.AccountDeletedAt, LastSeen: u.LastSeen,
		}
	}
	cacheMu.Unlock()
	catcher.Info(fmt.Sprintf("ad-audit: seeded cache with %d users", len(users)), nil)
}

func flushLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case <-ticker.C:
			flush()
		}
	}
}

func flush() {
	cacheMu.Lock()
	batch := make([]ingestUser, 0, len(dirty))
	ids := make([]string, 0, len(dirty))
	for id := range dirty {
		if cu := cache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			ids = append(ids, id)
		}
	}
	dirty = map[string]struct{}{} // cleared; events during the POST re-dirty independently
	cacheMu.Unlock()

	if len(batch) == 0 {
		return
	}
	if err := postBatch(batch); err != nil {
		_ = catcher.Error("ad-audit: flush failed; re-queuing", err, map[string]any{"count": len(batch)})
		cacheMu.Lock()
		for _, id := range ids {
			dirty[id] = struct{}{}
		}
		cacheMu.Unlock()
	}
}

func postBatch(users []ingestUser) error {
	body, err := json.Marshal(map[string]any{"users": users})
	if err != nil {
		return err
	}
	resp, err := request(http.MethodPost, "/api/v1/ad-audit/users", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ingest returned status %d", resp.StatusCode)
	}
	return nil
}
