package dto

import "github.com/utmstack/utmstack/backend/modules/compliance/domain"

type EnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type ControlResponse struct {
	domain.Control
	// System marks the shipped library, which is read-only: a tenant builds its
	// own controls beside it rather than forking it.
	System bool `json:"system"`
	Locked bool `json:"locked"`
}

type FrameworkResponse struct {
	domain.Framework
	System bool `json:"system"`
	Locked bool `json:"locked"`
}
