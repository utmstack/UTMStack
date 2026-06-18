package models

import (
	"time"

	"gorm.io/gorm"
)

// Recovery represents an imported YAML recovery directive.
// One row per yaml_id; source of truth is the YAML filesystem.
type Recovery struct {
	gorm.Model
	YamlID           string    `gorm:"type:varchar(128);not null;uniqueIndex"`
	YamlHash         string    `gorm:"type:char(64);not null"`
	Name             string    `gorm:"type:varchar(255);not null"`
	Description      string    `gorm:"type:text"`
	Shell            string    `gorm:"type:varchar(16);not null"`
	TargetOS         string    `gorm:"type:varchar(32);not null"`
	TargetVersionLte string    `gorm:"type:varchar(32)"`
	SuccessPattern   string    `gorm:"type:varchar(255);not null;default:'RECOVERY_OK'"`
	MaxConcurrency   int       `gorm:"not null;default:100"`
	MaxAttempts      int       `gorm:"not null;default:3"`
	RetryAfterSecs   int       `gorm:"type:int;not null;default:1800"`
	AckTimeoutSecs   int       `gorm:"type:int;not null;default:300"`
	ExpiresAt        time.Time `gorm:"not null"`
	DryRun           bool      `gorm:"not null;default:false"`
	Script           string    `gorm:"type:text;not null"`
	DependsOnYamlID  string    `gorm:"type:varchar(128)"`
	Status           string    `gorm:"type:varchar(16);not null;default:'ACTIVE';index:idx_recoveries_status"`
	BlockedReason    string    `gorm:"type:varchar(512)"`
	ImportedAt       time.Time `gorm:"not null;autoCreateTime"`
}

// RecoveryTarget represents one agent's execution slot for a given recovery.
// One row per (recovery_id, agent_id) pair.
type RecoveryTarget struct {
	gorm.Model
	RecoveryID    uint     `gorm:"not null;uniqueIndex:idx_recovery_target_unique;index:idx_target_recovery_status,priority:1"`
	Recovery      Recovery `gorm:"foreignKey:RecoveryID"`
	AgentID       uint     `gorm:"not null;uniqueIndex:idx_recovery_target_unique;index:idx_target_status_agent,priority:2"`
	Status        string   `gorm:"type:varchar(16);not null;default:'PENDING';index:idx_target_status_agent,priority:1;index:idx_target_recovery_status,priority:2"`
	Attempts      int      `gorm:"not null;default:0"`
	LastAttemptAt *time.Time
	LastCmdID     string `gorm:"type:varchar(64)"`
	LastResult    string `gorm:"type:text"`
	LastError     string `gorm:"type:varchar(512)"`
	CompletedAt   *time.Time
}
