package domain

import (
	"time"

	"github.com/google/uuid"
)

type Config struct {
	ID          uuid.UUID  `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID    uuid.UUID  `gorm:"column:tenant_id;type:uuid;not null;uniqueIndex:ux_app_config_tenant_key,priority:1"`
	Key         string     `gorm:"column:key;size:100;not null;uniqueIndex:ux_app_config_tenant_key,priority:2"`
	Label       string     `gorm:"column:label;size:150"`
	Description string     `gorm:"column:description"`
	Value       string     `gorm:"column:value"`
	IsSecret    bool       `gorm:"column:is_secret;not null;default:false"`
	UpdatedAt   *time.Time `gorm:"column:updated_at"`
	UpdatedBy   string     `gorm:"column:updated_by;size:254"`
}

func (Config) TableName() string { return "app_config" }
