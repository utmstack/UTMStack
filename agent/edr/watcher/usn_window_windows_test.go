//go:build windows

package watcher

import "testing"

func TestUSNJournalRetainedCursorWindow(t *testing.T) {
	journal := usnJournalData{FirstUsn: 10, LowestValidUsn: 20, NextUsn: 100}
	for _, tc := range []struct {
		next    int64
		outside bool
	}{{0, true}, {19, true}, {20, false}, {50, false}, {100, false}, {101, true}} {
		if got := outsideJournalWindow(tc.next, journal); got != tc.outside {
			t.Fatalf("cursor %d outside=%v want=%v", tc.next, got, tc.outside)
		}
	}
}
