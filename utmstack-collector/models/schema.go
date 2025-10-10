package models

import (
	"time"
)

type Container struct {
	ID      string            `json:"id"`
	Name    string            `json:"name"`
	Image   string            `json:"image"`
	Status  string            `json:"status"`
	State   string            `json:"state"`
	Created time.Time         `json:"created"`
	Labels  map[string]string `json:"labels"`
}

type LogEntry struct {
	ID            string `json:"id"`
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	Message       string `json:"message"`
}

type ContainerEvent struct {
	ID          string            `json:"id"`
	ContainerID string            `json:"container_id"`
	Action      string            `json:"action"` // start, stop, die, etc.
	Timestamp   time.Time         `json:"timestamp"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

type FilterRule struct {
	Type    string `yaml:"type"`    // image, label, name
	Pattern string `yaml:"pattern"` // regex pattern o string exacto
	Action  string `yaml:"action"`  // include, exclude
}

type ContainerFilter struct {
	rules []FilterRule
}

type Log struct {
	ID         string `gorm:"index"`
	CreatedAt  time.Time
	DataSource string
	Type       string
	Log        string
	Processed  bool
}
