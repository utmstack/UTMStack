package main

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const (
	auditdDedupTTL    = 5 * time.Minute
	auditdDedupCap    = 10_000
	auditdDedupEvictN = 1000
)

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

var (
	linuxCache = map[string]*cachedLinuxUser{}
	linuxDirty = map[string]struct{}{}

	hostnameMachineID = map[string]string{}

	resolveMu    sync.Mutex
	resolveQueue []resolveIntent

	dedupMu     sync.Mutex
	auditdDedup = map[string]time.Time{}
)

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
		n := flexString(cu.UIDNumber)
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
			cu.UIDNumber = string(*u.UIDNumber)
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
