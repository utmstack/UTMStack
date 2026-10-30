package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type IdentityAlias struct {
	ID          uuid.UUID                 `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TenantID    uuid.UUID                 `gorm:"column:tenant_id;size:36;not null;index:idx_identity_alias_tenant,priority:1" json:"tenantId"`
	Name        string                    `gorm:"column:name;size:64;not null" json:"name"`
	CreatedBy   string                    `gorm:"column:created_by;size:150;not null" json:"createdBy"`
	CreatedAt   time.Time                 `gorm:"column:created_at;not null;default:now()" json:"createdAt"`
	Assignments []IdentityAliasAssignment `gorm:"foreignKey:AliasID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (IdentityAlias) TableName() string { return "identity_alias" }

func ValidIdentityName(name string) bool {
	trimmed := strings.TrimSpace(name)
	runess := []rune(trimmed)
	return len(runess) >= 2 && len(runess) <= 64
}

type IdentityAliasAssignment struct {
	AliasID    uuid.UUID `gorm:"column:alias_id;type:uuid;primaryKey" json:"aliasId"`
	UserID     uuid.UUID `gorm:"column:user_id;type:uuid;primaryKey;uniqueIndex:idx_identity_alias_assignment_user" json:"userId"`
	AssignedBy string    `gorm:"column:assigned_by;size:150;not null" json:"assignedBy"`
	AssignedAt time.Time `gorm:"column:assigned_at;not null;default:now()" json:"assignedAt"`
}

func (IdentityAliasAssignment) TableName() string { return "identity_alias_assignment" }
