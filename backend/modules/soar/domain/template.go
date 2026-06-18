package domain

type AlertResponseActionTemplate struct {
	ID          int64  `gorm:"column:id;primaryKey"              json:"id,omitempty"`
	Title       string `gorm:"column:title;size:150;not null"    json:"title"`
	Description string `gorm:"column:description;type:text"      json:"description,omitempty"`
	Command     string `gorm:"column:command;type:text;not null" json:"command"`
	SystemOwner bool   `gorm:"column:system_owner;not null"      json:"systemOwner"`
}

func (AlertResponseActionTemplate) TableName() string { return "utm_response_action_template" }
