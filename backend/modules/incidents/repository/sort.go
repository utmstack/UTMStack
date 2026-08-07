package repository

import "strings"

func orderBy(sort string, sortable map[string]string, fallback string) string {
	field, dir, _ := strings.Cut(sort, ",")
	column, ok := sortable[strings.TrimSpace(field)]
	if !ok {
		return fallback
	}
	if strings.EqualFold(strings.TrimSpace(dir), "desc") {
		return column + " DESC"
	}
	return column + " ASC"
}

// The sortable columns of each list endpoint, keyed by the name the API uses.
var (
	incidentSortable = map[string]string{
		"incidentName":        "incident_name",
		"incidentStatus":      "incident_status",
		"incidentSeverity":    "incident_severity",
		"incidentAssignedTo":  "incident_assigned_to",
		"incidentCreatedDate": "incident_created_date",
	}

	historySortable = map[string]string{
		"action":            "action",
		"actionCreatedDate": "action_created_date",
		"actionCreatedBy":   "action_created_by",
	}
)
