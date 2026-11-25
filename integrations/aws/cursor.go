package main

import (
	"encoding/json"
	"sync"
)

// One goroutine polls per log stream, so a group's position is the merge of
// every stream's independently advancing NextForwardToken.
type cursorMap struct {
	mu      sync.Mutex
	entries map[string]*string
}

func newCursorMap() *cursorMap {
	return &cursorMap{entries: make(map[string]*string)}
}

// Only the stream's own goroutine may call this; a second writer for the
// same key breaks delete's invariant.
func (c *cursorMap) set(stream string, token *string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[stream] = token
}

func (c *cursorMap) get(stream string) (*string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	v, ok := c.entries[stream]
	return v, ok
}

// Callers must guarantee no set for this key can land afterwards, or the
// entry is resurrected: cancel the stream's goroutine first, or call this
// from that goroutine as it returns. To drop a live stream's token, use
// set(stream, nil) instead.
func (c *cursorMap) delete(stream string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, stream)
}

func (c *cursorMap) snapshot() map[string]*string {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]*string, len(c.entries))
	for k, v := range c.entries {
		out[k] = v
	}
	return out
}

func (c *cursorMap) marshalSnapshot() ([]byte, error) {
	return json.Marshal(c.snapshot())
}

// Only safe before the group's stream goroutines start.
func (c *cursorMap) replace(entries map[string]*string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if entries == nil {
		entries = make(map[string]*string)
	}
	c.entries = entries
}

// A returned token makes GetLogEvents ignore StartTime. On nil the caller
// must fall back to the group's baseline start time, not "now": streams are
// only discovered every 5 minutes, so "now" silently drops everything a
// late-discovered stream received before that.
func seedNextToken(cursors *cursorMap, stream string) *string {
	if token, ok := cursors.get(stream); ok {
		return token
	}
	return nil
}
