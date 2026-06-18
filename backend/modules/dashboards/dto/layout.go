package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type LayoutFilter struct {
	IDDashboard     *uint64 `form:"idDashboard"`
	IDVisualization *uint64 `form:"idVisualization"`
	database.Params
}
