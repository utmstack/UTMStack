package main

import "time"

// cursorKeyPrefix namespaces this plugin's keys in the shared bucket.
const cursorKeyPrefix = "sophos."

func sophosCursorKey(group *ModuleGroup) string {
	return cursorKeyPrefix + group.Key()
}

// cursorSnapshot is the persisted position for one group: StartTime is the
// event-time floor in Unix seconds (from_date), NextKey the pagination key
// (pageFromKey). buildURL sends only one per request, so the floor is the
// fallback when the key expires; both halves must always move together.
type cursorSnapshot struct {
	StartTime int64  `json:"startTime"`
	NextKey   string `json:"nextKey"`
}

// floor must be the job's own window start, never time.Now(): a clock-derived
// seed moves forward on every redelivery and silently skips whatever arrived
// between failed attempts.
func seedFrom(floor time.Time) cursorSnapshot {
	return cursorSnapshot{StartTime: floor.Unix(), NextKey: ""}
}

// floor must be the tick's own end, not the worker's clock: it is the earlier
// of the two, so a later fallback to from_date re-reads instead of skipping.
func (s cursorSnapshot) advanced(nextKey string, floor time.Time) cursorSnapshot {
	return cursorSnapshot{StartTime: floor.Unix(), NextKey: nextKey}
}

// A zero floor is a hazard, not a default: from_date=0 makes Sophos Central
// return its entire retained history.
func (s cursorSnapshot) usable() bool { return s.StartTime != 0 }
