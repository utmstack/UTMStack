package agent

import (
	"context"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const configReportInterval = 5 * time.Minute

var configAppliers = map[string]func(content string) error{
	"retention":   SetDataRetention,
	"update_hold": SetUpdateHold,
}

type configStateStore struct {
	mu    sync.Mutex
	state map[string]uint64 // key -> revision already applied
}

var configState = &configStateStore{state: make(map[string]uint64)}

var loadConfigStateOnce sync.Once

func ensureConfigStateLoaded() {
	loadConfigStateOnce.Do(loadConfigState)
}

func loadConfigState() {
	configState.mu.Lock()
	defer configState.mu.Unlock()

	var saved map[string]uint64
	if err := fs.ReadJSON(config.ConfigStateFile, &saved); err != nil || saved == nil {
		return
	}
	configState.state = saved
}

func currentConfigState() map[string]uint64 {
	configState.mu.Lock()
	defer configState.mu.Unlock()

	snapshot := make(map[string]uint64, len(configState.state))
	for k, v := range configState.state {
		snapshot[k] = v
	}
	return snapshot
}

func applyConfigUpdate(update *ConfigUpdate) {
	key := update.GetKey()
	apply, ok := configAppliers[key]
	if !ok {
		utils.Logger.LogF(100, "received config update for unrecognized key %q, ignoring", key)
		return
	}

	if err := apply(update.GetContent()); err != nil {
		utils.Logger.ErrorF("failed to apply config %q (revision %d): %v", key, update.GetRevision(), err)
		return
	}

	configState.mu.Lock()
	configState.state[key] = update.GetRevision()
	snapshot := make(map[string]uint64, len(configState.state))
	for k, v := range configState.state {
		snapshot[k] = v
	}
	configState.mu.Unlock()

	if err := fs.WriteJSON(config.ConfigStateFile, snapshot); err != nil {
		utils.Logger.ErrorF("failed to persist config state: %v", err)
	}

	utils.Logger.LogF(100, "applied config %q, now at revision %d", key, update.GetRevision())
}

func sendConfigStateReports(ctx context.Context, sender resultSender) {
	report := func() {
		err := sender.Send(&BidirectionalStream{
			StreamMessage: &BidirectionalStream_ConfigState{
				ConfigState: &ConfigState{Revisions: currentConfigState()},
			},
		})
		if err != nil {
			utils.Logger.LogF(100, "config state not delivered: %v", err)
		}
	}

	report()

	ticker := time.NewTicker(configReportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			report()
		}
	}
}
