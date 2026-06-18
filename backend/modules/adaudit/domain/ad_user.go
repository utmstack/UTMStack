package domain

import "time"

type ADUser struct {
	ID               uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         string     `gorm:"column:tenant_id;size:64;not null;uniqueIndex:idx_ad_user_tenant_sid" json:"tenantId"`
	SID              string     `gorm:"column:sid;size:128;not null;uniqueIndex:idx_ad_user_tenant_sid" json:"sid"`
	SamAccountName   string     `gorm:"column:sam_account_name;size:255" json:"samAccountName"`
	Domain           string     `gorm:"column:domain;size:255" json:"domain"`
	Active           bool       `gorm:"column:active;not null;default:true" json:"active"`
	AccountCreatedAt *time.Time `gorm:"column:account_created_at" json:"accountCreatedAt,omitempty"`
	LastLogon        *time.Time `gorm:"column:last_logon" json:"lastLogon,omitempty"`
	AccountDeletedAt *time.Time `gorm:"column:account_deleted_at" json:"accountDeletedAt,omitempty"`
	LastSeen         *time.Time `gorm:"column:last_seen" json:"lastSeen,omitempty"`
}

func (ADUser) TableName() string { return "ad_user" }
