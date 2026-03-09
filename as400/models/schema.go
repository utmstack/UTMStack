package models

import (
	"time"
)

type Log struct {
	ID         string `gorm:"index"`
	CreatedAt  time.Time
	DataSource string
	Type       string
	Log        string
	Processed  bool
}
