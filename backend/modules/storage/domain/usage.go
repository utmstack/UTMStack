package domain

import "time"

type Usage struct {
	Dataset   Dataset
	Documents int64
	Bytes     int64
	Oldest    time.Time
	Newest    time.Time
}

type Health struct {
	Status      string
	DiskUsedPct float64
	Message     string
}

const (
	HealthOK          = "ok"
	HealthDegraded    = "degraded"
	HealthUnavailable = "unavailable"
)
