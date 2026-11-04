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
	// Backoff deadlines persist here so a permanent error logs once per
	// window, not every tick (the processor is rebuilt each tick).
	UnavailableSources map[string]time.Time `json:"unavailableSources,omitempty"`
	// State diff for riskyUsers, which is a state collection, not an event
	// feed; the logs table has no id dedup, so only changes emit a record.
	RiskyUsers map[string]riskyUserState `json:"riskyUsers,omitempty"`
}

type riskyUserState struct {
	User         string `json:"user"`
	RiskLevel    string `json:"riskLevel"`
	RiskState    string `json:"riskState"`
	LastSeenTime time.Time `json:"lastSeen"`
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

func (c *cursorPayload) sourceUnavailable(source string) bool {
	t, ok := c.UnavailableSources[source]
	return ok && time.Now().UTC().Before(t)
}

func (c *cursorPayload) markSourceUnavailable(source string) {
	if c.UnavailableSources == nil {
		c.UnavailableSources = make(map[string]time.Time)
	}
	c.UnavailableSources[source] = time.Now().UTC().Add(time.Hour)
}
