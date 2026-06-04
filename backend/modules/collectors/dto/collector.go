package dto

import "time"

type CollectorListQuery struct {
	PageNumber int    `form:"pageNumber"`
	PageSize   int    `form:"pageSize"`
	Hostname   string `form:"hostname"`
	Module     string `form:"module"` // AS_400 | UTMSTACK | ""
	SortBy     string `form:"sortBy"`
}

type CollectorResponse struct {
	ID           int64      `json:"id"`
	Status       string     `json:"status"`
	CollectorKey string     `json:"collectorKey"`
	IP           string     `json:"ip"`
	Hostname     string     `json:"hostname"`
	Version      string     `json:"version"`
	Module       string     `json:"module"`
	LastSeen     *time.Time `json:"lastSeen"`
	GroupID      *int64     `json:"groupId"`
	Active       bool       `json:"active"`
}

type ListCollectorsResponse struct {
	Collectors []CollectorResponse `json:"collectors"`
	Total      int32               `json:"total"`
}
