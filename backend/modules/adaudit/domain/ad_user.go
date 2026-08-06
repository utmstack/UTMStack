package domain

import (
	"time"

	"github.com/google/uuid"
)

type ADUser struct {
	ID               uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID         uuid.UUID  `gorm:"column:tenant_id;size:36;not null;index:idx_ad_user_tenant_source,priority:1;uniqueIndex:idx_aduser_windows,priority:1;uniqueIndex:idx_aduser_linux,priority:1" json:"tenantId"`
	Source           string     `gorm:"column:source;size:16;not null;default:'windows';index:idx_ad_user_tenant_source,priority:2" json:"source"`
	SID              *string    `gorm:"column:sid;size:128;uniqueIndex:idx_aduser_windows,priority:2,where:source = 'windows'" json:"sid,omitempty"`
	SamAccountName   *string    `gorm:"column:sam_account_name;size:255" json:"samAccountName,omitempty"`
	Domain           *string    `gorm:"column:domain;size:255" json:"domain,omitempty"`
	MachineID        *string    `gorm:"column:machine_id;size:64;uniqueIndex:idx_aduser_linux,priority:2,where:source = 'linux'" json:"machineId,omitempty"`
	UIDNumber        *uint32    `gorm:"column:uid_number;uniqueIndex:idx_aduser_linux,priority:3" json:"uidNumber,omitempty"`
	Hostname         *string    `gorm:"column:hostname;size:255" json:"hostname,omitempty"`
	Username         *string    `gorm:"column:username;size:255" json:"username,omitempty"`
	Active           bool       `gorm:"column:active;not null" json:"active"`
	AccountCreatedAt *time.Time `gorm:"column:account_created_at" json:"accountCreatedAt,omitempty"`
	LastLogon        *time.Time `gorm:"column:last_logon" json:"lastLogon,omitempty"`
	AccountDeletedAt *time.Time `gorm:"column:account_deleted_at" json:"accountDeletedAt,omitempty"`
	LastSeen         *time.Time `gorm:"column:last_seen" json:"lastSeen,omitempty"`
}

func (ADUser) TableName() string { return "ad_user" }
