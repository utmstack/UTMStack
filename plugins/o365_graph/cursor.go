package main

import "time"

// cursorKeyPrefix namespaces this plugin's keys in the shared bucket.
const cursorKeyPrefix = "o365_graph."

func graphCursorKey(group *ModuleGroup) string {
	return cursorKeyPrefix + group.Key()
}

// windowEnd is the last boundary the consumer fully processed; sources maps
// each source name to its own position, because one source failing must not
// stall the others.
type cursorPayload struct {
	WindowEnd time.Time            `json:"windowEnd"`
	Sources   map[string]time.Time `json:"sources,omitempty"`
}

func (c *cursorPayload) sourcePosition(name string) time.Time {
	if c == nil || c.Sources == nil {
		return time.Time{}
	}
	return c.Sources[name]
}

func (c *cursorPayload) setSourcePosition(name string, t time.Time) {
	if c.Sources == nil {
		c.Sources = make(map[string]time.Time)
	}
	c.Sources[name] = t
}
