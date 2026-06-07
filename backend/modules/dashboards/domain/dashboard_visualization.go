package domain

type UtmDashboardVisualization struct {
	ID               uint64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IDVisualization  uint64  `gorm:"column:id_visualization;not null" json:"idVisualization"`
	IDDashboard      uint64  `gorm:"column:id_dashboard;not null" json:"idDashboard"`
	Order            int     `gorm:"column:dv_order;not null" json:"order"`
	Width            float64 `gorm:"column:dv_width;not null" json:"width"`
	Height           float64 `gorm:"column:dv_height;not null" json:"height"`
	Top              float64 `gorm:"column:dv_top;not null" json:"top"`
	Left             float64 `gorm:"column:dv_left;not null" json:"left"`
	ShowTimeFilter   *bool   `gorm:"column:dv_show_time_filter" json:"showTimeFilter,omitempty"`
	DefaultTimeRange *string `gorm:"column:dv_default_time_range" json:"defaultTimeRange,omitempty"`
	GridInfo         *string `gorm:"column:dv_grid_info" json:"gridInfo,omitempty"`
}

func (UtmDashboardVisualization) TableName() string { return "utm_dashboard_visualization" }
