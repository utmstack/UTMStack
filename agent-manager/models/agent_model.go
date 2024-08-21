package models

import (
	"time"

	"gorm.io/gorm"
)

type MalwareStatus string

const (
	New      MalwareStatus = "NEW"
	Deleted  MalwareStatus = "DELETED"
	Excluded MalwareStatus = "EXCLUDED"
	Restored MalwareStatus = "RESTORED"
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

type AgentMalwareDetection struct {
	gorm.Model
	AgentID     uint
	FilePath    string
	Sha256      string
	Md5         string
	Description string
	Status      MalwareStatus
}

type AgentMalwareHistory struct {
	gorm.Model
	MalwareId  uint
	PrevStatus MalwareStatus
	ToStatus   MalwareStatus
	ChangedBy  string
}

type AgentMalwareExclusion struct {
	gorm.Model
	AgentID            uint
	ExcludeFilePath    string
	ExcludedBy         string
	ExcludeDescription string
}

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
