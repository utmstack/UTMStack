package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const pluginName = "com.utmstack.ad-audit"

func main() {
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
	linuxDataType       = "linux"

	fEventCode  = "eventCode"
	fTargetUser = "eventDataTargetUserName"
	fTargetSID  = "eventDataTargetSid"
	fTargetSID2 = "eventDataTargetUserSid"
	fTargetDom  = "eventDataTargetDomainName"

	auditdDedupTTL    = 5 * time.Minute
	auditdDedupCap    = 10_000
	auditdDedupEvictN = 1000
)

func analyze(event *plugins.Event, _ plugins.Analysis_AnalyzeServer) error {
	switch event.GetDataType() {
	case wineventlogDataType:
		return handleWindows(event)
	case linuxDataType:
		return handleLinux(event)
	default:
		return io.EOF
	}
}

func handleWindows(event *plugins.Event) error {
	code := logEventCode(event)
	switch code {
	case evtUserCreated, evtUserDeleted, evtLogon:
	default:
		return io.EOF
	}

	name := targetUser(event)
	domain := targetDomain(event)
	sid := firstNonEmpty(logStr(event, fTargetSID), logStr(event, fTargetSID2))
	if isSystemAccount(name) || isSystemSID(sid) {
		return io.EOF
	}
	if sid == "" {
		if name == "" {
			return io.EOF
		}
		sid = strings.ToLower(strings.TrimSpace(domain + "\\" + name)) // stable fallback identity
	}

	applyEvent(event.GetTenantId(), sid, name, domain, code, parseTime(event.GetTimestamp()))
	return io.EOF
}

func handleLinux(event *plugins.Event) error {
	syslogID := logStr(event, "syslogIdentifier")
	switch syslogID {
	case "useradd", "userdel":
		return handleLinuxJournald(event, syslogID)
	}
	action := strings.TrimSpace(event.GetAction())
	if action == "" {
		action = logStr(event, "action")
	}
	switch action {
	case "USER_LOGIN", "USER_START":
		return handleLinuxAuditd(event, action)
	}
	return io.EOF
}

var journaldUserAddRe = regexp.MustCompile(`^new user:\s*name=([^,]+),\s*UID=(\d+)`)

var journaldUserDelRe = regexp.MustCompile(`^delete user\s+'([^']+)'`)

func handleLinuxJournald(event *plugins.Event, syslogID string) error {
	message := logStr(event, "message")
	if message == "" {
		return io.EOF
	}
	tenantID := event.GetTenantId()
	hostname := strings.TrimSpace(event.GetDataSource())
	machineID := logStr(event, "machineId")
	ts := parseTime(event.GetTimestamp())

	if machineID != "" && hostname != "" {
		key := tenantID + ":" + hostname
		cacheMu.Lock()
		if prev, ok := hostnameMachineID[key]; !ok || prev != machineID {
			hostnameMachineID[key] = machineID
			cacheMu.Unlock()
			enqueueResolveIntent(tenantID, hostname, machineID)
		} else {
			cacheMu.Unlock()
		}
	}

	switch syslogID {
	case "useradd":
		m := journaldUserAddRe.FindStringSubmatch(message)
		if len(m) < 3 {
			return io.EOF
		}
		username := strings.TrimSpace(m[1])
		uid := strings.TrimSpace(m[2])
		if !isHumanUID(uid) || username == "" {
			return io.EOF
		}
		applyLinuxEvent(tenantID, machineID, hostname, uid, username, "create", ts)

	case "userdel":
		m := journaldUserDelRe.FindStringSubmatch(message)
		if len(m) < 2 {
			return io.EOF
		}
		username := strings.TrimSpace(m[1])
		if username == "" {
			return io.EOF
		}
		applyLinuxEvent(tenantID, machineID, hostname, "", username, "delete", ts)
	}
	return io.EOF
}

func handleLinuxAuditd(event *plugins.Event, action string) error {
	tenantID := event.GetTenantId()
	hostname := strings.TrimSpace(event.GetDataSource())
	ts := parseTime(event.GetTimestamp())

	sequence := logNumStr(event, "sequence")
	if dedupCheckAndMark(tenantID, hostname, sequence) {
		return io.EOF
	}

	var subobj string
	switch action {
	case "USER_LOGIN":
		subobj = "userlogin"
	case "USER_START":
		subobj = "userstart"
	default:
		return io.EOF
	}

	result := logStrPath(event, subobj, "result")
	if result != "success" {
		return io.EOF
	}

	auid := logStrPath(event, subobj, "auid")
	if auid == "unset" {
		return io.EOF
	}

	acct := logStrPath(event, subobj, "acct")
	if acct == "(invalid user)" {
		return io.EOF
	}
	if isSystemAccount(acct) {
		return io.EOF
	}

	id := logStrPath(event, subobj, "id")

	var machineID string
	if hostname != "" {
		cacheMu.Lock()
		machineID = hostnameMachineID[tenantID+":"+hostname]
		cacheMu.Unlock()
	}

	switch action {
	case "USER_LOGIN":
		if !isHumanUID(id) {
			return io.EOF
		}
		username := acct
		if username == "" {
			username = lookupLinuxUsername(tenantID, machineID, hostname, id)
		}
		if username == "" {
			return io.EOF
		}
		applyLinuxEvent(tenantID, machineID, hostname, id, username, "login", ts)

	case "USER_START":
		if acct == "" {
			return io.EOF
		}
		uidForApply := ""
		if isHumanUID(id) {
			uidForApply = id
		}
		applyLinuxEvent(tenantID, machineID, hostname, uidForApply, acct, "session", ts)
	}
	return io.EOF
}

func dedupCheckAndMark(tenantID, hostname, sequence string) bool {
	if hostname == "" || sequence == "" {
		return false
	}
	key := tenantID + ":" + hostname + ":" + sequence

	dedupMu.Lock()
	defer dedupMu.Unlock()

	now := time.Now()
	if seen, ok := auditdDedup[key]; ok {
		if now.Sub(seen) < auditdDedupTTL {
			return true
		}
	}

	auditdDedup[key] = now

	if len(auditdDedup) > auditdDedupCap {
		for k, ts := range auditdDedup {
			if now.Sub(ts) >= auditdDedupTTL {
				delete(auditdDedup, k)
			}
		}
		if len(auditdDedup) > auditdDedupCap {
			evictOldestN(auditdDedup, auditdDedupEvictN)
		}
	}
	return false
}

func evictOldestN(m map[string]time.Time, n int) {
	type kt struct {
		k string
		t time.Time
	}
	all := make([]kt, 0, len(m))
	for k, t := range m {
		all = append(all, kt{k, t})
	}
	sort.Slice(all, func(i, j int) bool { return all[i].t.Before(all[j].t) })
	if n > len(all) {
		n = len(all)
	}
	for i := 0; i < n; i++ {
		delete(m, all[i].k)
	}
}

func logStrPath(event *plugins.Event, path ...string) string {
	log := event.GetLog()
	if log == nil || len(path) == 0 {
		return ""
	}
	first := log[path[0]]
	if first == nil {
		return ""
	}
	v := first.AsInterface()
	for i := 1; i < len(path); i++ {
		m, ok := v.(map[string]interface{})
		if !ok {
			return ""
		}
		v = m[path[i]]
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}

func lookupLinuxUsername(tenantID, machineID, hostname, uid string) string {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	if machineID != "" && uid != "" {
		key := tenantID + ":linux:resolved:" + machineID + ":" + uid
		if cu, ok := linuxCache[key]; ok {
			return cu.Username
		}
	}
	prefix := tenantID + ":linux:provisional:" + hostname + ":"
	for k, cu := range linuxCache {
		if strings.HasPrefix(k, prefix) && cu.UIDNumber == uid {
			return cu.Username
		}
	}
	return ""
}

func isHumanUID(uid string) bool {
	if uid == "" || uid == "unset" {
		return false
	}
	n, err := strconv.Atoi(uid)
	if err != nil {
		return false
	}
	return n >= 1000
}

func applyLinuxEvent(tenantID, machineID, hostname, uid, username, eventType string, et *time.Time) {
	if username == "" && uid == "" {
		return
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	var key string
	if machineID != "" && uid != "" {
		key = tenantID + ":linux:resolved:" + machineID + ":" + uid
	} else if hostname != "" && username != "" {
		key = tenantID + ":linux:provisional:" + hostname + ":" + username
	} else {
		return
	}

	cu := linuxCache[key]
	if cu == nil {
		cu = &cachedLinuxUser{
			TenantID:  tenantID,
			MachineID: machineID,
			Hostname:  hostname,
			UIDNumber: uid,
			Username:  username,
			Active:    true,
		}
		linuxCache[key] = cu
	}
	before := *cu

	if machineID != "" && cu.MachineID == "" {
		cu.MachineID = machineID
	}
	if uid != "" && cu.UIDNumber == "" {
		cu.UIDNumber = uid
	}
	if username != "" && cu.Username == "" {
		cu.Username = username
	}
	if hostname != "" && cu.Hostname == "" {
		cu.Hostname = hostname
	}

	cu.LastSeen = laterTime(cu.LastSeen, et)
	switch eventType {
	case "create":
		cu.AccountCreatedAt = firstSet(cu.AccountCreatedAt, et)
		cu.Active = true
	case "delete":
		cu.Active = false
		cu.AccountDeletedAt = et
	case "login":
		cu.LastLogon = laterTime(cu.LastLogon, et)
		cu.Active = true
	case "session":
		cu.Active = true
	}

	if *cu != before {
		linuxDirty[key] = struct{}{}
	}
}

func logStr(event *plugins.Event, key string) string {
	log := event.GetLog()
	if log == nil || log[key] == nil {
		return ""
	}
	return strings.TrimSpace(log[key].GetStringValue())
}

func logNumStr(event *plugins.Event, key string) string {
	log := event.GetLog()
	if log == nil || log[key] == nil {
		return ""
	}
	switch v := log[key].AsInterface().(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func logEventCode(event *plugins.Event) string {
	log := event.GetLog()
	if log == nil || log[fEventCode] == nil {
		return ""
	}
	switch v := log[fEventCode].AsInterface().(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return ""
	}
}

func targetUser(event *plugins.Event) string {
	if u := strings.TrimSpace(event.GetTarget().GetUser()); u != "" {
		return u
	}
	return logStr(event, fTargetUser)
}

func targetDomain(event *plugins.Event) string {
	if d := strings.TrimSpace(event.GetTarget().GetDomain()); d != "" {
		return d
	}
	return logStr(event, fTargetDom)
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

func isSystemSID(sid string) bool {
	switch s := strings.ToUpper(strings.TrimSpace(sid)); {
	case s == "":
		return false
	case s == "S-1-5-18", s == "S-1-5-19", s == "S-1-5-20": // LocalSystem, LocalService, NetworkService
		return true
	case s == "S-1-5-7", s == "S-1-0-0": // Anonymous Logon, Null SID
		return true
	case strings.HasPrefix(s, "S-1-5-90-0-"), strings.HasPrefix(s, "S-1-5-96-0-"): // DWM-*, UMFD-*
		return true
	}
	return false
}

// ── In-memory cache ───────────────────────────────────────────────────────────

type cachedWindowsUser struct {
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

type cachedLinuxUser struct {
	TenantID         string
	MachineID        string
	Hostname         string
	UIDNumber        string
	Username         string
	Active           bool
	AccountCreatedAt *time.Time
	LastLogon        *time.Time
	AccountDeletedAt *time.Time
	LastSeen         *time.Time
}

type resolveIntent struct {
	TenantID  string
	Hostname  string
	MachineID string
}

var (
	cacheMu      sync.Mutex
	windowsCache = map[string]*cachedWindowsUser{}
	windowsDirty = map[string]struct{}{}

	linuxCache = map[string]*cachedLinuxUser{}
	linuxDirty = map[string]struct{}{}

	hostnameMachineID = map[string]string{}

	resolveMu    sync.Mutex
	resolveQueue []resolveIntent

	dedupMu     sync.Mutex
	auditdDedup = map[string]time.Time{}
)

func applyEvent(tenantID, sid, name, domain, code string, et *time.Time) {
	id := tenantID + ":" + sid

	cacheMu.Lock()
	defer cacheMu.Unlock()

	cu := windowsCache[id]
	if cu == nil {
		cu = &cachedWindowsUser{TenantID: tenantID, SID: sid, Active: true}
		windowsCache[id] = cu
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
		windowsDirty[id] = struct{}{}
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
	Source           string     `json:"source,omitempty"`
	SID              string     `json:"sid,omitempty"`
	SamAccountName   string     `json:"samAccountName,omitempty"`
	Domain           string     `json:"domain,omitempty"`
	MachineID        *string    `json:"machineId,omitempty"`
	UIDNumber        *string    `json:"uidNumber,omitempty"`
	Hostname         *string    `json:"hostname,omitempty"`
	Username         *string    `json:"username,omitempty"`
	Active           *bool      `json:"active"`
	AccountCreatedAt *time.Time `json:"accountCreatedAt"`
	LastLogon        *time.Time `json:"lastLogon"`
	AccountDeletedAt *time.Time `json:"accountDeletedAt"`
	LastSeen         *time.Time `json:"lastSeen"`
}

func (cu *cachedWindowsUser) toIngest() ingestUser {
	active := cu.Active
	return ingestUser{
		TenantID:         cu.TenantID,
		Source:           "windows",
		SID:              cu.SID,
		SamAccountName:   cu.SamAccountName,
		Domain:           cu.Domain,
		Active:           &active,
		AccountCreatedAt: cu.AccountCreatedAt,
		LastLogon:        cu.LastLogon,
		AccountDeletedAt: cu.AccountDeletedAt,
		LastSeen:         cu.LastSeen,
	}
}

func (cu *cachedLinuxUser) toIngest() ingestUser {
	active := cu.Active
	u := ingestUser{
		Source:           "linux",
		TenantID:         cu.TenantID,
		Active:           &active,
		AccountCreatedAt: cu.AccountCreatedAt,
		LastLogon:        cu.LastLogon,
		AccountDeletedAt: cu.AccountDeletedAt,
		LastSeen:         cu.LastSeen,
	}
	if cu.MachineID != "" {
		m := cu.MachineID
		u.MachineID = &m
	}
	if cu.UIDNumber != "" {
		n := cu.UIDNumber
		u.UIDNumber = &n
	}
	if cu.Hostname != "" {
		h := cu.Hostname
		u.Hostname = &h
	}
	if cu.Username != "" {
		n := cu.Username
		u.Username = &n
	}
	return u
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
	seedWindows()
	seedLinux()
}

func seedWindows() {
	if backendURL == "" {
		return
	}
	resp, err := request(http.MethodGet, "/api/v1/ad-audit/users/sync?source=windows", nil)
	if err != nil {
		_ = catcher.Error("ad-audit: seed windows request failed", err, nil)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = catcher.Error("ad-audit: seed windows returned non-200", fmt.Errorf("status %d", resp.StatusCode), nil)
		return
	}

	users, err := decodeSeed(resp.Body, "windows")
	if err != nil {
		return
	}
	cacheMu.Lock()
	for _, u := range users {
		windowsCache[u.TenantID+":"+u.SID] = &cachedWindowsUser{
			TenantID:         u.TenantID,
			SID:              u.SID,
			SamAccountName:   u.SamAccountName,
			Domain:           u.Domain,
			Active:           u.Active == nil || *u.Active,
			AccountCreatedAt: u.AccountCreatedAt,
			LastLogon:        u.LastLogon,
			AccountDeletedAt: u.AccountDeletedAt,
			LastSeen:         u.LastSeen,
		}
	}
	cacheMu.Unlock()
	catcher.Info(fmt.Sprintf("ad-audit: seeded windows cache with %d users", len(users)), nil)
}

func seedLinux() {
	if backendURL == "" {
		return
	}
	resp, err := request(http.MethodGet, "/api/v1/ad-audit/users/sync?source=linux", nil)
	if err != nil {
		_ = catcher.Error("ad-audit: seed linux request failed", err, nil)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = catcher.Error("ad-audit: seed linux returned non-200", fmt.Errorf("status %d", resp.StatusCode), nil)
		return
	}
	users, err := decodeSeed(resp.Body, "linux")
	if err != nil {
		return
	}
	cacheMu.Lock()
	for _, u := range users {
		cu := &cachedLinuxUser{
			TenantID:         u.TenantID,
			Active:           u.Active == nil || *u.Active,
			AccountCreatedAt: u.AccountCreatedAt,
			LastLogon:        u.LastLogon,
			AccountDeletedAt: u.AccountDeletedAt,
			LastSeen:         u.LastSeen,
		}
		if u.MachineID != nil {
			cu.MachineID = *u.MachineID
		}
		if u.UIDNumber != nil {
			cu.UIDNumber = *u.UIDNumber
		}
		if u.Hostname != nil {
			cu.Hostname = *u.Hostname
		}
		if u.Username != nil {
			cu.Username = *u.Username
		}

		if cu.MachineID != "" && cu.Hostname != "" {
			hostnameMachineID[cu.TenantID+":"+cu.Hostname] = cu.MachineID
		}

		var key string
		if cu.MachineID != "" && cu.UIDNumber != "" {
			key = cu.TenantID + ":linux:resolved:" + cu.MachineID + ":" + cu.UIDNumber
		} else if cu.Hostname != "" && cu.Username != "" {
			key = cu.TenantID + ":linux:provisional:" + cu.Hostname + ":" + cu.Username
		} else {
			continue
		}
		linuxCache[key] = cu
	}
	cacheMu.Unlock()
	catcher.Info(fmt.Sprintf("ad-audit: seeded linux cache with %d users", len(users)), nil)
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
	batch := make([]ingestUser, 0, len(windowsDirty)+len(linuxDirty))
	windowsIDs := make([]string, 0, len(windowsDirty))
	linuxIDs := make([]string, 0, len(linuxDirty))

	for id := range windowsDirty {
		if cu := windowsCache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			windowsIDs = append(windowsIDs, id)
		}
	}
	for id := range linuxDirty {
		if cu := linuxCache[id]; cu != nil {
			batch = append(batch, cu.toIngest())
			linuxIDs = append(linuxIDs, id)
		}
	}
	windowsDirty = map[string]struct{}{}
	linuxDirty = map[string]struct{}{}
	cacheMu.Unlock()

	if len(batch) > 0 {
		if err := postBatch(batch); err != nil {
			_ = catcher.Error("ad-audit: flush failed; re-queuing", err, map[string]any{"count": len(batch)})
			cacheMu.Lock()
			for _, id := range windowsIDs {
				windowsDirty[id] = struct{}{}
			}
			for _, id := range linuxIDs {
				linuxDirty[id] = struct{}{}
			}
			cacheMu.Unlock()
		}
	}

	resolveMu.Lock()
	intents := resolveQueue
	resolveQueue = nil
	resolveMu.Unlock()

	for _, intent := range intents {
		if err := resolveLinuxIdentity(intent.TenantID, intent.Hostname, intent.MachineID); err != nil {
			_ = catcher.Error("ad-audit: resolve failed", err, map[string]any{
				"hostname":  intent.Hostname,
				"machineID": intent.MachineID,
			})
			continue
		}
		cacheMu.Lock()
		provisionalPrefix := intent.TenantID + ":linux:provisional:" + intent.Hostname + ":"
		for k, cu := range linuxCache {
			if !strings.HasPrefix(k, provisionalPrefix) {
				continue
			}
			if cu.UIDNumber == "" {
				continue
			}
			newKey := intent.TenantID + ":linux:resolved:" + intent.MachineID + ":" + cu.UIDNumber
			if _, exists := linuxCache[newKey]; exists {
				delete(linuxCache, k)
				continue
			}
			cu.MachineID = intent.MachineID
			linuxCache[newKey] = cu
			delete(linuxCache, k)
			linuxDirty[newKey] = struct{}{}
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

func enqueueResolveIntent(tenantID, hostname, machineID string) {
	resolveMu.Lock()
	resolveQueue = append(resolveQueue, resolveIntent{
		TenantID:  tenantID,
		Hostname:  hostname,
		MachineID: machineID,
	})
	resolveMu.Unlock()
}

var resolveLinuxIdentityFn = resolveLinuxIdentityDefault

func resolveLinuxIdentity(tenantID, hostname, machineID string) error {
	return resolveLinuxIdentityFn(tenantID, hostname, machineID)
}

func resolveLinuxIdentityDefault(tenantID, hostname, machineID string) error {
	if backendURL == "" || tenantID == "" || hostname == "" || machineID == "" {
		return nil
	}
	body, err := json.Marshal(map[string]string{
		"tenant_id":  tenantID,
		"hostname":   hostname,
		"machine_id": machineID,
	})
	if err != nil {
		return err
	}
	resp, err := request(http.MethodPost, "/api/v1/ad-audit/users/resolve", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("resolve returned status %d", resp.StatusCode)
	}
	return nil
}

func decodeSeed(r io.Reader, source string) ([]ingestUser, error) {
	var users []ingestUser

	dec := json.NewDecoder(r)
	for {
		var u ingestUser
		err := dec.Decode(&u)
		if errors.Is(err, io.EOF) {
			return users, nil
		}
		if err != nil {
			_ = catcher.Error("ad-audit: decoding the seed stream failed", err, map[string]any{"source": source})
			return nil, err
		}
		users = append(users, u)
	}
}
