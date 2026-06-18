package domain

import "time"

type User struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement"`
	Login            string `gorm:"size:50;uniqueIndex;not null"`
	PasswordHash     string `gorm:"size:60;not null"`
	FirstName        string `gorm:"size:50"`
	LastName         string `gorm:"size:50"`
	Email            string `gorm:"size:254;uniqueIndex"`
	Activated        bool   `gorm:"not null;default:false"`
	LangKey          string `gorm:"size:6"`
	ImageURL         string `gorm:"size:256;column:image_url"`
	ActivationKey    string `gorm:"size:20"`
	ResetKey         string `gorm:"size:20"`
	ResetDate        *time.Time
	TFASecret        string `gorm:"column:tfa_secret;size:100"`
	TFAMethod        string `gorm:"column:tfa_method;size:50"`
	DefaultPassword  bool   `gorm:"default:false"`
	FSManager        bool   `gorm:"column:fs_manager"`
	CreatedBy        string `gorm:"size:50;not null"`
	CreatedDate      time.Time
	LastModifiedBy   string `gorm:"size:50"`
	LastModifiedDate *time.Time
}

func (User) TableName() string { return "jhi_user" }
