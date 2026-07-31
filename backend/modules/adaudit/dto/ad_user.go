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
	Search   string `form:"search"`   // substring on samAccountName/sid
	TenantID string `form:"tenantId"` // exact
	Source   string `form:"source"`   // "windows"|"linux"|"" (all)
	Active   *bool  `form:"active"`
	Status   string `form:"status"` // active|disabled|deleted|stale|service — overrides Active when set
	Sort     string `form:"sort"`   // recent (last_seen desc) | name (default, samAccountName asc)
	database.Params
}

// ADUserStatsQuery scopes the inventory aggregation to a single tenant (optional).
type ADUserStatsQuery struct {
	TenantID string `form:"tenantId"`
}

type DomainCount struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}

// SourceCount is the by-source breakdown returned by GET /ad-audit/stats.
type SourceCount struct {
	Windows int64 `json:"windows"`
	Linux   int64 `json:"linux"`
}

// ADUserStats is the inventory roll-up the UI overview renders. Counts honor the
// optional tenant scope; Tenants is always the global distinct list so the
// tenant picker stays stable regardless of the active scope.
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
	Tenants  []string      `json:"tenants"`
}

// ResolveLinuxIdentityRequest is the payload the ad-audit plugin sends when it
// learns the machine-id for a host that already has provisional Linux user rows.
// The backend updates all matching provisional rows to set machine_id: if a
// resolved row already exists for (tenant_id, machine_id, uid_number),
// the provisional row is left untouched.
type ResolveLinuxIdentityRequest struct {
	TenantID  string `json:"tenant_id" binding:"required"`
	Hostname  string `json:"hostname"  binding:"required"`
	MachineID string `json:"machine_id" binding:"required"`
}
