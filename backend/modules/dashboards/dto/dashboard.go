package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type DashboardFilter struct {
	Name string `form:"name"`
	database.Params
}
