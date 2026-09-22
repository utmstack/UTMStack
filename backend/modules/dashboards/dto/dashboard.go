package dto

import "github.com/utmstack/utmstack/backend/pkg/database"

type DashboardFilter struct {
	Name string `form:"name"`
	// Dismissed, when true, returns only dashboards the user has removed (for
	// a "restore a default" picker). Omitted or false returns the normal,
	// visible set — dismissed dashboards never show up there.
	Dismissed *bool `form:"dismissed"`
	database.Params
}
