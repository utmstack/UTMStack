package repository

import (
	"fmt"
	"strings"
)

func scopedSQL(query, logsTable, alertsTable string, page, size int) string {
	if page < 1 {
		page = 1
	}
	q := strings.TrimSpace(strings.TrimRight(strings.TrimSpace(query), ";"))

	return fmt.Sprintf(
		"WITH logs AS (SELECT * FROM %s WHERE tenantId = ?), alerts AS (SELECT * FROM %s WHERE tenantId = ?) "+
			"SELECT * FROM (%s) LIMIT %d OFFSET %d",
		logsTable, alertsTable, q, size, (page-1)*size,
	)
}
