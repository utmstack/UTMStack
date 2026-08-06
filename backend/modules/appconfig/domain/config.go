package domain

import (
	"strings"
	"time"
)

type Config struct {
	ID                   int64  `gorm:"primaryKey"`
	TenantID             string `gorm:"column:tenant_id;size:36;not null;default:'';uniqueIndex:idx_config_tenant_key,priority:1"`
	ConfParamShort       string `gorm:"size:100;not null;uniqueIndex:idx_config_tenant_key,priority:2"`
	ConfParamLarge       string
	ConfParamDescription string
	ConfParamValue       string
	ConfParamRegexp      string
	ConfParamRequired    *bool
	ConfParamDatatype    string // "password" => secret
	ConfParamOption      string
	ModificationTime     *time.Time
	ModificationUser     string
}

func (Config) TableName() string { return "app_config" }

func (c Config) IsSecret() bool { return strings.EqualFold(c.ConfParamDatatype, "password") }
func (c *Config) SetIsSecret()  { c.ConfParamDatatype = "password" }
