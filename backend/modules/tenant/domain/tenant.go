package domain

import (
	"time"

	"github.com/google/uuid"
)

type TenantStatus string

const (
	StatusActive     TenantStatus = "ACTIVE"
	StatusSuspended  TenantStatus = "SUSPENDED"
	StatusTerminated TenantStatus = "TERMINATED"
)

type SupportAccess string

const (
	SupportNone SupportAccess = "NONE"
	SupportRead SupportAccess = "READ"
	SupportFull SupportAccess = "FULL"
)

type Tenant struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Name          string        `gorm:"size:255;not null" json:"name"`
	Domain        string        `gorm:"size:253;not null;uniqueIndex" json:"domain"`
	Status        TenantStatus  `gorm:"size:16;not null;default:ACTIVE;index" json:"status"`
	SupportAccess SupportAccess `gorm:"size:8;not null;default:NONE" json:"supportAccess"`
	Limits        Limits        `gorm:"embedded;embeddedPrefix:limit_" json:"limits"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type Limits struct {
	MaxAIRequests *int `gorm:"column:max_ai_requests" json:"maxAIRequests"`
}

func (l Limits) Valid() bool {
	return l.MaxAIRequests == nil || *l.MaxAIRequests >= 0
}

// AllowanceOf resolves what this tenant may actually spend, given what the
// instance itself is allowed. A negative instance limit means the instance has
// no cap of its own — and then the tenant's cap is the only bound there is,
// which is the whole point of an operational limit. Comparing against it as a
// number would make every tenant cap larger than -1, so all of them would be
// thrown away and an operator handing out "none" would be handing out
// everything.
func (l Limits) AllowanceOf(instanceLimit int) int {
	if l.MaxAIRequests == nil {
		return instanceLimit
	}
	if instanceLimit < 0 {
		return *l.MaxAIRequests
	}
	// A tenant is never handed more of the instance's quota than it has.
	if *l.MaxAIRequests > instanceLimit {
		return instanceLimit
	}
	return *l.MaxAIRequests
}

func (Tenant) TableName() string { return "tenant" }
