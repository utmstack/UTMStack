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
	database.Params
}
