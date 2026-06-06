package domain

type UtmLogstashFilterGroup struct {
	ID               int64   `gorm:"primaryKey;autoIncrement"                                        json:"id"`
	GroupName        string  `gorm:"column:group_name;size:150;not null;uniqueIndex"    json:"groupName"`
	GroupDescription *string `gorm:"column:group_description"                          json:"groupDescription"`
	SystemOwner      bool    `gorm:"column:system_owner;not null;default:false"         json:"systemOwner"`
}

func (UtmLogstashFilterGroup) TableName() string {
	return "utm_logstash_filter_group"
}
