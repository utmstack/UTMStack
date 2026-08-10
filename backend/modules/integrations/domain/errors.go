package domain

import "errors"

var (
	ErrIntegrationNotFound     = errors.New("integration not found")
	ErrSystemIntegration       = errors.New("system integrations cannot be modified or deleted")
	ErrInvalidIngestType       = errors.New("invalid ingest type")
	ErrConfigGroupNotFound     = errors.New("configuration group not found")
	ErrNotConfigurable         = errors.New("integration is not configurable (no schema)")
	ErrRequiredConfigEmpty     = errors.New("required config field is empty")
	ErrInvalidCollectorConfig  = errors.New("invalid collector data-type configuration")
	ErrAgentManagerUnavailable = errors.New("agent manager is not available")
)
