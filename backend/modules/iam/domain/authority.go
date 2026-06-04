package domain

type Permission struct {
	ID          uint64 `gorm:"primaryKey;autoIncrement"`
	Name        string `gorm:"size:100;uniqueIndex;not null"`
	Description string `gorm:"size:500"`
	Resource    string `gorm:"size:100;not null"`
	Action      string `gorm:"size:50;not null"`
}

func (Permission) TableName() string { return "permissions" }

// Authority is a role, backed by the legacy jhi_authority table — the name is
// its identity (no numeric id). name_show holds the display name.
type Authority struct {
	Name        string `gorm:"column:name;primaryKey;size:50"`
	NameShow    string `gorm:"column:name_show;size:100"`
	Description string `gorm:"column:description;size:500"`
}

func (Authority) TableName() string { return "jhi_authority" }

type UserAuthority struct {
	UserID        uint64 `gorm:"column:user_id;primaryKey"`
	AuthorityName string `gorm:"column:authority_name;primaryKey;size:50"`
}

func (UserAuthority) TableName() string { return "jhi_user_authority" }

type AuthorityPermission struct {
	AuthorityName string `gorm:"column:authority_name;primaryKey;size:50"`
	PermissionID  uint64 `gorm:"column:permission_id;primaryKey"`
}

func (AuthorityPermission) TableName() string { return "authority_permissions" }
