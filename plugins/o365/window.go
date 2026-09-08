package main

import "time"

const (
	// The Management Activity API returns partial results, not an error, beyond 24h; 12h leaves boundary margin.
	maxCollectionWindow = 12 * time.Hour
	// Clamp short of the API's 7 day retention: the request is issued seconds after now is captured.
	maxCollectionLookback = 6 * 24 * time.Hour
)

func nextWindow(position, now time.Time, size time.Duration) (time.Time, time.Time, bool) {
	if size <= 0 || !position.Before(now) {
		return time.Time{}, time.Time{}, false
	}

	end := position.Add(size)
	if end.After(now) {
		end = now
	}

	return position, end, true
}
