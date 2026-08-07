package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/pkg/database"
)

type IngestUser struct {
	TenantID         string     `json:"tenantId"`
	Source           string     `json:"source,omitempty"` // "windows"|"linux"; defaults to "windows" if absent
	SID              string     `json:"sid,omitempty"`    // required when source=windows; enforced in usecase
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

type IngestRequest struct {
	Users []IngestUser `json:"users" binding:"required,min=1"`
}

type ADUserFilter struct {
	Search string `form:"search"` // substring on samAccountName/sid
	Source string `form:"source"` // "windows"|"linux"|"" (all)
	Active *bool  `form:"active"`
	Status string `form:"status"` // active|disabled|deleted|stale|service — overrides Active when set
	Sort   string `form:"sort"`   // recent (last_seen desc) | name (default, samAccountName asc)
	database.Params
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

type SourceCount struct {
	Windows int64 `json:"windows"`
	Linux   int64 `json:"linux"`
}

type ADUserStats struct {
	Total    int64         `json:"total"`
	Active   int64         `json:"active"`
	Disabled int64         `json:"disabled"`
	Deleted  int64         `json:"deleted"`
	Stale    int64         `json:"stale"`
	Service  int64         `json:"service"`
	Seen24h  int64         `json:"seen_24h"`
	BySource SourceCount   `json:"by_source"`
	ByDomain []DomainCount `json:"by_domain"`
}

type ResolveLinuxIdentityRequest struct {
	TenantID  string `json:"tenant_id" binding:"required"`
	Hostname  string `json:"hostname"  binding:"required"`
	MachineID string `json:"machine_id" binding:"required"`
}
