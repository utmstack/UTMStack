package domain

import (
	"strings"
	"time"
)

type Config struct {
	ID                   int64  `gorm:"primaryKey"`
	ConfParamShort       string `gorm:"size:100;not null"`
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

func (Config) TableName() string { return "utm_configuration_parameter" }

func (c Config) IsSecret() bool { return strings.EqualFold(c.ConfParamDatatype, "password") }
func (c *Config) SetIsSecret()  { c.ConfParamDatatype = "password" }
