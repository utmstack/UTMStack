package usecase

import (
	"fmt"
	"regexp"
)

func ApplyPagination(sql string, page, size int) string {
	hasLimit := reHasLimit.MatchString(sql)
	hasOffset := reHasOffset.MatchString(sql)

	offset := (page - 1) * size

	switch {
	case hasLimit && hasOffset:
		return sql
	case hasLimit && !hasOffset:
		return fmt.Sprintf("%s OFFSET %d", sql, offset)
	case !hasLimit && hasOffset:
		return fmt.Sprintf("%s LIMIT %d", sql, size)
	default:
		return fmt.Sprintf("%s LIMIT %d OFFSET %d", sql, size, offset)
	}
}

var (
	reHasLimit  = regexp.MustCompile(`(?i)\bLIMIT\b`)
	reHasOffset = regexp.MustCompile(`(?i)\bOFFSET\b`)
)
