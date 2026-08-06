package domain

import (
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	Name        string `gorm:"column:name;primaryKey;size:100"`
	Description string `gorm:"size:500"`
}

func (Permission) TableName() string { return "permissions" }

type RolePermission struct {
	RoleID         uuid.UUID `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionName string    `gorm:"column:permission_name;primaryKey;size:100"`
}

func (RolePermission) TableName() string { return "role_permission" }

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:ux_role_tenant_name,priority:1"`
	Name        string    `gorm:"size:50;not null;uniqueIndex:ux_role_tenant_name,priority:2"`
	DisplayName string    `gorm:"column:display_name;size:100"`
	Description string    `gorm:"size:500"`
	SystemOwner bool      `gorm:"not null;default:false"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Role) TableName() string { return "role" }

func (Role) SystemFlagColumn() string { return "system_owner" }

type UserRole struct {
	UserID   uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey"`
	RoleID   uuid.UUID `gorm:"column:role_id;type:uuid;primaryKey"`
	TenantID uuid.UUID `gorm:"column:tenant_id;type:uuid;not null;index"`
}

func (UserRole) TableName() string { return "user_role" }
