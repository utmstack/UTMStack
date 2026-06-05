package domain

type UtmIndexPattern struct {
	ID            int64   `gorm:"primaryKey;autoIncrement"                        json:"id"`
	Pattern       string  `gorm:"column:pattern;size:100;not null;uniqueIndex"     json:"pattern"`
	PatternModule *string `gorm:"column:pattern_module;size:500"                   json:"patternModule"`
	PatternSystem *bool   `gorm:"column:pattern_system"                            json:"patternSystem"`
	IsActive      *bool   `gorm:"column:is_active"                                 json:"isActive"`
}

func (UtmIndexPattern) TableName() string { return "utm_index_pattern" }
