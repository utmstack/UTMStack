package domain

type DashboardVisualization struct {
	ID              uint64 `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	IDDashboard     uint64 `gorm:"column:id_dashboard;not null;index" json:"idDashboard"`
	IDVisualization uint64 `gorm:"column:id_visualization;not null" json:"idVisualization"`
	Layout          string `gorm:"column:layout" json:"layout"`
}

func (DashboardVisualization) TableName() string { return "dashboard_visualization" }
