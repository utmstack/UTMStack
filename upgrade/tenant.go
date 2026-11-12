package upgrade

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	defaultTenantIDStr = "ce66672c-e36d-4761-a8c8-90058fee1a24"
	defaultTenantName  = "UTMStack"
)

var defaultTenantID = uuid.MustParse(defaultTenantIDStr)

// tenantRow is the v12 tenant table shape, re-declared here so the upgrade
// module has no dependency on the backend.
type tenantRow struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name          string    `gorm:"size:255;not null"`
	Domain        string    `gorm:"size:253;not null;uniqueIndex"`
	Status        string    `gorm:"size:16;not null;default:ACTIVE;index"`
	SupportAccess string    `gorm:"size:8;not null;default:NONE"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (tenantRow) TableName() string { return "tenant" }

func ensureDefaultTenant(db *gorm.DB, domain string) error {
	tenant := tenantRow{
		ID:            defaultTenantID,
		Name:          defaultTenantName,
		Domain:        domain,
		Status:        "ACTIVE",
		SupportAccess: "NONE",
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tenant).Error
}
