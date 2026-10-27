package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const (
	evtUserCreated = "4720" // A user account was created
	evtUserDeleted = "4726" // A user account was deleted
	evtLogon       = "4624" // An account was successfully logged on

	fEventCode  = "eventCode"
	fTargetUser = "eventDataTargetUserName"
	fTargetSID  = "eventDataTargetSid"
	fTargetSID2 = "eventDataTargetUserSid"
	fTargetDom  = "eventDataTargetDomainName"
)

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

var (
	windowsCache = map[string]*cachedWindowsUser{}
	windowsDirty = map[string]struct{}{}
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
