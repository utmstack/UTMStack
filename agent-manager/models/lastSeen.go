package models

import "time"

type LastSeen struct {
	ConnectorType string
	ConnectorID   uint
	LastPing      time.Time
}
