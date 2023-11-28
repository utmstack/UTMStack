package models

import (
	"gorm.io/gorm"
	"time"
)

type AgentStatus int32

const (
	Online  AgentStatus = 0
	Offline AgentStatus = 1
	Unknown AgentStatus = 2
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
	AgentKey       string               `gorm:"type:string;index"`
	Commands       []AgentCommand       `gorm:"foreignKey:AgentID"`
	AgentConfigs   []AgentConfiguration `gorm:"foreignKey:AgentID"`
	AgentModules   []AgentConfiguration `gorm:"foreignKey:AgentID"`
	DeletedAt      *time.Time           `gorm:"uniqueIndex:idx_hostname_deleted;index:idx_agent_delete"`
	RegisterBy     string               `gorm:"not null"`
	DeletedBy      string               `gorm:""`
	AgentTypeID    *uint
	AgentType      *AgentType `gorm:"foreignKey:AgentTypeID"`
	AgentGroupID   *uint
	AgentGroup     *AgentGroup `gorm:"foreignKey:AgentGroupID"`
	Mac            string
	OsMajorVersion string
	OsMinorVersion string
	Aliases        string
	Addresses      string
}

type AgentLastSeen struct {
	AgentKey string    `gorm:"primaryKey"`
	LastPing time.Time `gorm:"index"`
}

type AgentCommand struct {
	gorm.Model
	AgentID       uint
	Command       string
	CommandStatus AgentCommandStatus
	Result        string
	Agent         Agent  `gorm:"foreignKey:AgentID"`
	ExecutedBy    string `gorm:"not null"`
	CmdId         string `gorm:"not null"`
	OriginType    string `gorm:"not null"`
	OriginId      string `gorm:"not null"`
	Reason        string `gorm:"not null"`
}

type AgentConfiguration struct {
	gorm.Model
	AgentID         uint
	ConfKey         string
	ConfValue       string
	ConfName        string
	ConfDescription string
	ConfDatatype    string
	ConfRequired    bool
	ConfRegex       string
}

type AgentGroup struct {
	gorm.Model
	Id               uint
	GroupName        string `gorm:"uniqueIndex:idx_group_deleted;not null"`
	GroupDescription string
	DeletedAt        *time.Time `gorm:"uniqueIndex:idx_group_deleted"`
}

type AgentType struct {
	gorm.Model
	TypeName string
}
