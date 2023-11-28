package models

import "gorm.io/gorm"

type AgentModule struct {
	gorm.Model
	AgentID       uint
	ShortName     string
	LargeName     string
	Description   string
	Enabled       bool
	AllowDisabled bool
	ModuleConfigs []AgentModuleConfiguration `gorm:"foreignKey:AgentModuleID"`
}
type AgentModuleConfiguration struct {
	AgentModuleID   uint
	ShortName       string
	ConfKey         string
	ConfValue       string
	ConfName        string
	ConfDescription string
	ConfDatatype    string
	ConfRequired    bool
	ConfRegex       string
}
