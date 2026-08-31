package main

import "github.com/utmstack/UTMStack/plugins/shared/identity"

// LogRecord is one ingested o365 audit record with its deterministic identity.
type LogRecord struct {
	ID  string
	Raw string
}

// recordID is Microsoft's own per-record GUID, so the same source event always
// hashes to the same identity, making a re-read recognisable. Nothing collapses
// duplicates today: utmstack.logs ORDER BY excludes id, and alert dedup is
// opt-in per rule. utmTenantId is redundant with groupKey but kept explicit.
func eventIdentity(utmTenantId, groupKey, contentUri, recordID string) string {
	return identity.Hash(utmTenantId, groupKey, contentUri, recordID)
}
