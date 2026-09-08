package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
)

const stateFileName = "o365_state.json"

type persistedState struct {
	Groups map[string]persistedGroup `json:"groups"`
}

type persistedGroup struct {
	GroupName string    `json:"groupName"`
	WindowEnd time.Time `json:"windowEnd"`
}

func defaultStatePath() string {
	return filepath.Join(plugins.WorkDir, "pipeline", stateFileName)
}

func loadPositions(path string) map[int32]persistedGroup {
	positions := make(map[int32]persistedGroup)

	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			_ = catcher.Error("unable to read the collection state, resuming from the default window", err, map[string]any{
				"process": "plugin_com.utmstack.o365",
				"file":    path,
			})
		}
		return positions
	}

	var state persistedState
	// A corrupt checkpoint degrades to the default seeding instead of stopping collection.
	if err := json.Unmarshal(data, &state); err != nil {
		_ = catcher.Error("unable to parse the collection state, resuming from the default window", err, map[string]any{
			"process": "plugin_com.utmstack.o365",
			"file":    path,
		})
		return positions
	}

	for key, group := range state.Groups {
		id, err := strconv.ParseInt(key, 10, 32)
		if err != nil {
			continue
		}
		positions[int32(id)] = group
	}

	return positions
}

func savePositions(path string, positions map[int32]persistedGroup) {
	state := persistedState{Groups: make(map[string]persistedGroup, len(positions))}
	for id, group := range positions {
		state.Groups[strconv.FormatInt(int64(id), 10)] = group
	}

	data, err := json.Marshal(state)
	if err != nil {
		_ = catcher.Error("unable to encode the collection state", err, map[string]any{
			"process": "plugin_com.utmstack.o365",
		})
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		_ = catcher.Error("unable to create the collection state directory", err, map[string]any{
			"process": "plugin_com.utmstack.o365",
			"file":    path,
		})
		return
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		_ = catcher.Error("unable to write the collection state", err, map[string]any{
			"process": "plugin_com.utmstack.o365",
			"file":    path,
		})
		return
	}

	if err := os.Rename(tmp, path); err != nil {
		_ = catcher.Error("unable to finalize the collection state", err, map[string]any{
			"process": "plugin_com.utmstack.o365",
			"file":    path,
		})
	}
}
