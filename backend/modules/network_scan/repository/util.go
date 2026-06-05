package repository

import (
	"strings"

	"github.com/utmstack/utmstack/backend/modules/network_scan/domain"
)

const (
	defaultPage = 1
	defaultSize = 20
	maxSize     = 200
)

func normalizePage(p domain.Pagination) (page, size int) {
	page = p.Page
	if page < 1 {
		page = defaultPage
	}
	size = p.PageSize
	if size < 1 {
		size = defaultSize
	}
	if size > maxSize {
		size = maxSize
	}
	return page, size
}

// parseSort accepts the Spring-style "field,direction" tokens and returns a safe ORDER BY clause.
// Empty string → empty result. Only allow whitelisted columns to avoid injection.
func parseSort(raw string, allowed map[string]string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parts := strings.SplitN(raw, ",", 2)
	col, ok := allowed[strings.TrimSpace(parts[0])]
	if !ok {
		return ""
	}
	dir := "ASC"
	if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[1]), "desc") {
		dir = "DESC"
	}
	return col + " " + dir
}
