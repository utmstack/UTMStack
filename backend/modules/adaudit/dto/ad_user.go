package dto

import (
	"time"

	"github.com/utmstack/utmstack/backend/pkg/database"
)

type IngestUser struct {
	TenantID         string     `json:"tenantId"`
	SID              string     `json:"sid" binding:"required"`
	SamAccountName   string     `json:"samAccountName"`
	Domain           string     `json:"domain"`
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
	ByDomain []DomainCount `json:"by_domain"`
	Tenants  []string      `json:"tenants"`
}
