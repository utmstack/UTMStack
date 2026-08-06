package domain

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus string

const (
	UserStatusPending   UserStatus = "pending"
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID           uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;index;uniqueIndex:ux_user_tenant_email,priority:1"`
	Email              string     `gorm:"size:254;not null;uniqueIndex:ux_user_tenant_email,priority:2"`
	Name               string     `gorm:"size:100"`
	PasswordHash       *string    `gorm:"size:60"`
	Status             UserStatus `gorm:"size:16;not null;default:'pending';index"`
	IdentityProviderID *uuid.UUID `gorm:"column:identity_provider_id;type:uuid;index"`
	LangKey            string     `gorm:"size:6"`
	ImageURL           string     `gorm:"size:256;column:image_url"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

func (User) TableName() string { return "user" }
