package main

import (
	"encoding/json"

	"github.com/utmstack/UTMStack/plugins/shared/identity"
)

// LogRecord is one ingested Sophos Central event with its deterministic identity.
type LogRecord struct {
	ID  string
	Raw string
}

// Marshaling and identifying stay in one function so Raw is exactly the bytes
// eventIdentity hashed; computing them separately lets the two diverge.
func newLogRecord(groupKey string, item map[string]any) (LogRecord, error) {
	raw, err := json.Marshal(item)
	if err != nil {
		return LogRecord{}, err
	}

	return LogRecord{
		ID:  eventIdentity(groupKey, item, raw),
		Raw: string(raw),
	}, nil
}

// eventIdentity gives the same source event the same id on every ingest, so a
// re-read event is recognisable. Nothing downstream collapses it: utmstack.logs
// does not order by id, and alert dedup is opt-in per rule. The "id"/"raw"
// prefix keeps the two hash spaces disjoint.
func eventIdentity(groupKey string, item map[string]any, raw []byte) string {
	if nativeID, ok := item["id"].(string); ok && nativeID != "" {
		return identity.Hash("id", groupKey, nativeID)
	}
	return identity.Hash("raw", groupKey, string(raw))
}
