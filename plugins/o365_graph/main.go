package main

import (
	"context"
	"runtime"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/plugins/shared/coordination"
)

var (
	activeGroupsMu sync.RWMutex
	activeGroups   = make(map[string]*ModuleGroup)
)

const coordinationRetryDelay = 5 * time.Second

func main() {
	mode := plugins.GetCfg(processName).Env.Mode
	if mode != "worker" {
		return
	}

	go StartConfigurationSystem()

	for i := 0; i < 2*runtime.NumCPU(); i++ {
		go plugins.SendLogsFromChannel(pluginName)
	}

	ctx := context.Background()

	// Blocks until NATS coordination is up. Fails closed: never ingests without
	// coordination.
	setup, err := coordination.SetupWithRetry(ctx, coordination.DialQueuePath(processName, queuePlugin), coordinationRetryDelay, func(err error) {
		_ = catcher.Error("coordination setup failed, waiting to retry", err, map[string]any{
			"process": processName,
		})
	})
	if err != nil {
		_ = catcher.Error("coordination setup cancelled before it could succeed", err, map[string]any{"process": processName})
		return
	}
	defer setup.Close()
	logCoordinationReady()

	go runQueueConsumer(ctx, setup.Consumer, setup.Cursors, getEncryptionKey)

	watchConfigAndPull(setup.Scheduler, setup.Publisher, uuid.New().String())
}

func syncActiveGroups(newConfig *ConfigurationSection) {
	activeGroupsMu.Lock()
	defer activeGroupsMu.Unlock()

	if newConfig == nil || !newConfig.ModuleActive {
		catcher.Info("Module deactivated, clearing all groups", map[string]any{
			"process": processName,
		})
		activeGroups = make(map[string]*ModuleGroup)
		return
	}

	newGroups := make(map[string]*ModuleGroup)
	for _, grp := range newConfig.ModuleGroups {
		newGroups[grp.Key()] = grp
	}

	for key := range activeGroups {
		if _, exists := newGroups[key]; !exists {
			catcher.Info("Group removed from configuration", map[string]any{
				"group":   key,
				"process": processName,
			})
		}
	}

	for key := range newGroups {
		if _, exists := activeGroups[key]; !exists {
			catcher.Info("New group added to configuration", map[string]any{
				"group":   key,
				"process": processName,
			})
		}
	}

	activeGroups = newGroups
}

func getActiveGroups() []*ModuleGroup {
	activeGroupsMu.RLock()
	defer activeGroupsMu.RUnlock()

	groups := make([]*ModuleGroup, 0, len(activeGroups))
	for _, grp := range activeGroups {
		groups = append(groups, grp)
	}
	return groups
}

// The local startTime below only seeds published jobs; authoritative per-group
// progress lives in the persisted cursor, since another worker may handle a tick.
func watchConfigAndPull(scheduler coordination.Store, publisher coordination.JobPublisher, holder string) {
	time.Sleep(3 * time.Second)

	initialConfig := GetConfig()
	if initialConfig != nil && initialConfig.ModuleActive {
		syncActiveGroups(initialConfig)
	}

	delay := 5 * time.Minute
	ticker := time.NewTicker(delay)
	defer ticker.Stop()

	startTime := time.Now().UTC().Add(-delay)

	for {
		select {
		case newConfig := <-GetConfigUpdateChannel():
			catcher.Info("Received config update, syncing groups", map[string]any{
				"moduleActive": newConfig != nil && newConfig.ModuleActive,
				"process":      processName,
			})
			syncActiveGroups(newConfig)

		case <-ticker.C:
			endTime := time.Now().UTC()

			groups := getActiveGroups()
			if len(groups) == 0 {
				catcher.Info("No active groups, skipping pull", map[string]any{
					"process": processName,
				})
				startTime = endTime.Add(1 * time.Nanosecond)
				continue
			}

			publishTickJobs(context.Background(), scheduler, publisher, holder, groups, startTime, endTime)

			startTime = endTime.Add(1 * time.Nanosecond)
		}
	}
}
