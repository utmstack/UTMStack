package models

import (
	"time"

	"gorm.io/gorm"
)

type AgentCommandStatus int32

const (
	Queue    AgentCommandStatus = 1
	Pending  AgentCommandStatus = 2
	Executed AgentCommandStatus = 3
	Error    AgentCommandStatus = 4
)

type Agent struct {
	gorm.Model
	Ip             string
	Hostname       string `gorm:"uniqueIndex:idx_hostname_deleted;not null"`
	Os             string
	Platform       string
	Version        string
	AgentKey       string     `gorm:"type:string;index"`
	DeletedAt      *time.Time `gorm:"uniqueIndex:idx_hostname_deleted;index:idx_agent_delete"`
	RegisterBy     string     `gorm:"not null"`
	DeletedBy      string
	Mac            string
	OsMajorVersion string
	OsMinorVersion string
	Aliases        string
	Addresses      string
}

type AgentCommand struct {
	gorm.Model
	AgentID       uint
	Command       string
	CommandStatus AgentCommandStatus
	Result        string
	ExecutedBy    string `gorm:"not null"`
	CmdId         string `gorm:"not null"`
	OriginType    string `gorm:"not null"`
	OriginId      string `gorm:"not null"`
	Reason        string `gorm:"not null"`
}
