package main

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/google/uuid"
	"github.com/utmstack/config-client-go/types"
)

var timeGroups = make(map[int]int)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

func pullLogs(group types.ModuleGroup) {
	epoch := int(time.Now().Unix())

	_, ok := timeGroups[group.ModuleID]
	if !ok {
		timeGroups[group.ModuleID] = epoch - delayCheck
	}

	defer func() {
		timeGroups[group.ModuleID] = epoch + 1
	}()

	agent := getSophosCentralProcessor(group)
	logs, err := agent.getLogs(timeGroups[group.ModuleID])
	if err != nil {
		_ = catcher.Error("error getting logs", err, map[string]any{})
		return
	}

	if len(logs) > 0 {
		for _, log := range logs {
			plugins.EnqueueLog(&plugins.Log{
				Id:         uuid.New().String(),
				TenantId:   defaultTenant,
				DataType:   "sophos-central",
				DataSource: group.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        log,
			})
		}
	}
}
