package models

import "time"

type LastSeen struct {
	ConnectorType string
	ID            uint
	LastPing      time.Time
}
