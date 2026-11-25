package models

import "github.com/threatwinds/go-sdk/entities"

// EntityIngestResult is an entity that was not accepted, and why.
type EntityIngestResult struct {
	Entity *entities.Entity
	Err    error
	Status int
}

// IngestOutcome says what happened to a batch of entities.
type IngestOutcome struct {
	// OKCount is how many entities were accepted.
	OKCount int
	// Sent are the entities that were accepted.
	Sent []*entities.Entity
	// Skipped is how many were not sent because they had already been.
	Skipped int
	// Failures are the entities that were attempted and not accepted.
	Failures []EntityIngestResult
	// ProvisioningBlocked is set when the instance is not provisioned for ingestion, which makes
	// every further attempt fail the same way.
	ProvisioningBlocked bool
}
