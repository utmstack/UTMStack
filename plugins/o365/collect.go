package main

import (
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/UTMStack/plugins/o365/config"
)

type windowPositions struct {
	mu   sync.Mutex
	path string
	at   map[int32]persistedGroup
}

func newWindowPositions(path string) *windowPositions {
	return &windowPositions{path: path, at: loadPositions(path)}
}

func (w *windowPositions) positionFor(groupID int32, seed time.Time) time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()

	position, seen := w.at[groupID]
	if !seen {
		w.at[groupID] = persistedGroup{WindowEnd: seed}
		return seed
	}
	return position.WindowEnd
}

func (w *windowPositions) advanceTo(group *config.ModuleGroup, at time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.at[group.Id] = persistedGroup{GroupName: group.GroupName, WindowEnd: at}
	savePositions(w.path, w.at)
}

func (w *windowPositions) retain(groupIDs map[int32]struct{}) {
	// An empty list means the configuration has not loaded yet, so pruning would recreate the gap.
	if len(groupIDs) == 0 {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	dropped := false
	for id := range w.at {
		if _, configured := groupIDs[id]; !configured {
			delete(w.at, id)
			dropped = true
		}
	}

	if dropped {
		savePositions(w.path, w.at)
	}
}

type pullFunc func(startTime, endTime time.Time, group *config.ModuleGroup) error

func collectGroup(positions *windowPositions, group *config.ModuleGroup, now, seed time.Time, doPull pullFunc) {
	position := positions.positionFor(group.Id, seed)

	if oldest := now.Add(-maxCollectionLookback); position.Before(oldest) {
		_ = catcher.Error("collection position was older than the lookback limit, that much backlog was skipped", nil, map[string]any{
			"process":   "plugin_com.utmstack.o365",
			"group":     group.GroupName,
			"skipped":   oldest.Sub(position).String(),
			"from":      position.Format(time.RFC3339),
			"resumedAt": oldest.Format(time.RFC3339),
		})

		position = oldest
		positions.advanceTo(group, oldest)
	}

	start, end, ok := nextWindow(position, now, maxCollectionWindow)
	if !ok {
		return
	}

	// Leaving the position at a failed window's start is what makes that range retryable.
	if err := doPull(start, end, group); err != nil {
		return
	}

	positions.advanceTo(group, end)
}
