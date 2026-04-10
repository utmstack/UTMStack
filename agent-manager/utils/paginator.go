package utils

import (
	"strings"

	"gorm.io/gorm"
)

type Pagination struct {
	Limit  int    `json:"limit,omitempty"`
	Page   int    `json:"page,omitempty"`
	Sort   string `json:"sort,omitempty"`
	Offset int    `json:"-"`
}

type Page[T any] struct {
	Total int64 `json:"total"`
	Rows  []T   `json:"rows"`
}

func NewPaginator(limit int, page int, sort string) Pagination {
	p := Pagination{
		Limit: limit,
		Page:  page,
	}
	p.Offset = (p.Page - 1) * p.Limit

	if len(sort) > 0 {
		srt := make([]string, 0)
		for _, s := range strings.Split(sort, "&") {
			parts := strings.SplitN(s, ",", 2)
			field := parts[0]
			if !IsValidFieldName(field) {
				continue
			}
			direction := "asc"
			if len(parts) == 2 {
				d := strings.ToLower(strings.TrimSpace(parts[1]))
				if d == "desc" {
					direction = "desc"
				}
			}
			srt = append(srt, field+" "+direction)
		}
		p.Sort = strings.Join(srt, ",")
	}
	return p
}

func (p *Pagination) PagingScope(db *gorm.DB) *gorm.DB {
	return db.Limit(p.Limit).Offset(p.Offset).Order(p.Sort)
}
