package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

// cursorSnapshot is the payload persisted per unit. StartTime is the event-time
// floor in Unix milliseconds; it must be persisted with the offsets, not
// recomputed on activation, or resuming filters out the backlog they exist for.
type cursorSnapshot struct {
	StartTime uint64            `json:"startTime"`
	Offsets   map[string]uint64 `json:"offsets"`
}

type cursorState struct {
	mu        sync.Mutex
	startTime uint64
	offsets   map[string]uint64
}

// newCursorState returns an unseeded state; only activateCursor may seed it.
func newCursorState() *cursorState {
	return &cursorState{offsets: make(map[string]uint64)}
}

func (c *cursorState) seedFromNow(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startTime = uint64(now.UnixMilli())
	c.offsets = make(map[string]uint64)
}

func (c *cursorState) restore(s cursorSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.startTime = s.StartTime
	c.offsets = make(map[string]uint64, len(s.Offsets))
	for feed, offset := range s.Offsets {
		c.offsets[feed] = offset
	}
}

func (c *cursorState) setOffset(feed string, offset uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.offsets[feed] = offset
}

// offset returns the feed's resume point, zero if never consumed. Zero makes
// gofalcon omit the offset parameter entirely, safe only because the event-time
// floor still bounds what the feed then delivers.
func (c *cursorState) offset(feed string) uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.offsets[feed]
}

// startsAfter returns the event-time floor in Unix milliseconds.
func (c *cursorState) startsAfter() uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.startTime
}

func (c *cursorState) snapshot() cursorSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	out := cursorSnapshot{StartTime: c.startTime, Offsets: make(map[string]uint64, len(c.offsets))}
	for feed, offset := range c.offsets {
		out.Offsets[feed] = offset
	}
	return out
}

// marshalSnapshot implements coordination.CursorSnapshot.
func (c *cursorState) marshalSnapshot() ([]byte, error) {
	return json.Marshal(c.snapshot())
}

// activateCursor establishes a newly owned unit's position and returns the
// revision its heartbeat must continue from. A fresh seed is persisted before
// any event is consumed, so a crash before the first heartbeat does not make the
// successor reseed and skip whatever arrived in between.
func activateCursor(ctx context.Context, cursors coordination.CursorStore, key string, state *cursorState, now time.Time) (uint64, error) {
	var persisted cursorSnapshot

	rev, found, err := coordination.LoadCursorInto(ctx, cursors, key, &persisted)
	if err != nil {
		// Fail closed: seeding from now would look like a clean start while
		// silently skipping everything since the last understood position.
		return rev, err
	}

	// A zero StartTime is unusable, not adoptable: it would emit every event
	// CrowdStrike still retains.
	if found && persisted.StartTime != 0 {
		state.restore(persisted)
		return rev, nil
	}

	state.seedFromNow(now)

	data, err := state.marshalSnapshot()
	if err != nil {
		return rev, err
	}

	saved, err := cursors.Save(ctx, key, coordination.Cursor{Data: data, Revision: rev})
	if err != nil {
		// The unit still runs: the floor is set in memory, so nothing historical
		// is emitted. Only durability is delayed to the first heartbeat.
		return rev, err
	}

	return saved.Revision, nil
}
