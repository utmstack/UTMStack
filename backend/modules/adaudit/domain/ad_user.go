package domain

import "time"

type ADUser struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         string     `gorm:"column:tenant_id;size:64;not null" json:"tenantId"`
	Source           string     `gorm:"column:source;size:16;not null;default:'windows'" json:"source"`
	SID              *string    `gorm:"column:sid;size:128" json:"sid,omitempty"`
	SamAccountName   string     `gorm:"column:sam_account_name;size:255" json:"samAccountName,omitempty"`
	Domain           string     `gorm:"column:domain;size:255" json:"domain,omitempty"`
	MachineID        *string    `gorm:"column:machine_id;size:64" json:"machineId,omitempty"`
	UIDNumber        *string    `gorm:"column:uid_number;size:32" json:"uidNumber,omitempty"`
	Hostname         *string    `gorm:"column:hostname;size:255" json:"hostname,omitempty"`
	Username         *string    `gorm:"column:username;size:255" json:"username,omitempty"`
	Active           bool       `gorm:"column:active;not null;default:true" json:"active"`
	AccountCreatedAt *time.Time `gorm:"column:account_created_at" json:"accountCreatedAt,omitempty"`
	LastLogon        *time.Time `gorm:"column:last_logon" json:"lastLogon,omitempty"`
	AccountDeletedAt *time.Time `gorm:"column:account_deleted_at" json:"accountDeletedAt,omitempty"`
	LastSeen         *time.Time `gorm:"column:last_seen" json:"lastSeen,omitempty"`
}

func (ADUser) TableName() string { return "ad_user" }
