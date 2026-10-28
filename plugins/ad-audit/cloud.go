package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

type cachedCloudUser struct {
	TenantID         string
	Source           string
	UPN              string
	DisplayName      string
	Domain           string
	Active           bool
	AccountCreatedAt *time.Time
	LastLogon        *time.Time
	AccountDeletedAt *time.Time
	LastSeen         *time.Time
}

var (
	cloudCache = map[string]*cachedCloudUser{}
	cloudDirty = map[string]struct{}{}
)

var (
	userOpCreate   = map[string]bool{"Add user": true, "Create user": true}
	userOpDelete   = map[string]bool{"Delete user": true, "Delete a user": true}
	userOpDisabled = map[string]bool{"Set user as disabled": true}
	userOpEnabled  = map[string]bool{"Set user as enabled": true}
)

func strField(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

func graphTime(v any) *time.Time {
	s, _ := v.(string)
	if s == "" {
		return nil
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.UTC); err == nil {
			return &t
		}
	}
	return nil
}

func applyCloudEvent(tenantID, source, upn, displayName, domain, eventType string, et *time.Time) {
	if tenantID == "" || upn == "" || source == "" {
		return
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	key := tenantID + ":" + source + ":" + upn
	cu := cloudCache[key]
	if cu == nil {
		cu = &cachedCloudUser{TenantID: tenantID, Source: source, UPN: upn, Active: true}
		cloudCache[key] = cu
	}
	before := *cu

	if displayName != "" {
		cu.DisplayName = displayName
	}
	if domain != "" {
		cu.Domain = domain
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
	case "disabled":
		cu.Active = false
	case "enabled":
		cu.Active = true
	}

	if *cu != before {
		cloudDirty[key] = struct{}{}
	}
}

func errNon200(status int) error { return fmt.Errorf("status %d", status) }

var cloudSources = []string{"o365", "azure", "gcp", "aws"}

func seedCloud() {
	if backendURL == "" {
		return
	}
	for _, source := range cloudSources {
		resp, err := request(http.MethodGet, "/api/v1/ad-audit/users/sync?source="+source, nil)
		if err != nil {
			_ = catcher.Error("ad-audit: seed cloud request failed", err, map[string]any{"source": source})
			continue
		}
		if resp.StatusCode != http.StatusOK {
			_ = catcher.Error("ad-audit: seed cloud returned non-200", errNon200(resp.StatusCode), map[string]any{"source": source})
			resp.Body.Close()
			continue
		}
		users, err := decodeSeed(resp.Body, source)
		resp.Body.Close()
		if err != nil {
			continue
		}
		cacheMu.Lock()
		for _, u := range users {
			if u.Username == nil || *u.Username == "" {
				continue
			}
			upn := *u.Username
			cu := &cachedCloudUser{
				TenantID:         u.TenantID,
				Source:           source,
				UPN:              upn,
				DisplayName:      u.SamAccountName,
				Active:           u.Active == nil || *u.Active,
				AccountCreatedAt: u.AccountCreatedAt,
				LastLogon:        u.LastLogon,
				AccountDeletedAt: u.AccountDeletedAt,
				LastSeen:         u.LastSeen,
			}
			if u.Domain != "" {
				cu.Domain = u.Domain
			}
			cloudCache[u.TenantID+":"+source+":"+upn] = cu
		}
		seeded := len(cloudCache)
		cacheMu.Unlock()
		catcher.Info("ad-audit: seeded cloud cache", map[string]any{"users": seeded, "source": source})
	}
}

func (cu *cachedCloudUser) toIngest() ingestUser {
	active := cu.Active
	u := ingestUser{
		Source:           cu.Source,
		TenantID:         cu.TenantID,
		SamAccountName:   cu.DisplayName,
		Active:           &active,
		AccountCreatedAt: cu.AccountCreatedAt,
		LastLogon:        cu.LastLogon,
		AccountDeletedAt: cu.AccountDeletedAt,
		LastSeen:         cu.LastSeen,
	}
	un := cu.UPN
	u.Username = &un
	if cu.Domain != "" {
		u.Domain = cu.Domain
	}
	return u
}
