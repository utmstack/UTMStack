package main

import "time"

// cursorKeyPrefix namespaces this plugin's keys in the shared bucket.
const cursorKeyPrefix = "o365."

func o365CursorKey(group *ModuleGroup) string {
	return cursorKeyPrefix + group.Key()
}

// Boundary of the last successfully ingested window. CursorStore treats
// Cursor.Data as opaque bytes and never parses this.
type cursorPayload struct {
	WindowEnd time.Time `json:"windowEnd"`
}
