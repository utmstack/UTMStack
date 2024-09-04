package processor

import (
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/config-client-go/types"
)

var timeGroups = make(map[int]int)

const delayCheck = 300
const DefaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

func PullLogs(group types.ModuleGroup) {
	go_sdk.Logger().Info("starting log sync for : %s", group.GroupName)

	epoch := int(time.Now().Unix())

	_, ok := timeGroups[group.ModuleID]
	if !ok {
		timeGroups[group.ModuleID] = epoch - delayCheck
	}

	defer func() {
		timeGroups[group.ModuleID] = epoch + 1
	}()

	agent := GetSophosCentralProcessor(group)
	logs, err := agent.GetLogs(group, timeGroups[group.ModuleID])
	if err != nil {
		go_sdk.Logger().ErrorF("error getting logs: %v", err)
		return
	}

	if len(logs) > 0 {
		for _, log := range logs {
			LogQueue <- &go_sdk.Log{
				Id:         uuid.New().String(),
				TenantId:   DefaultTenant,
				DataType:   "sophos-central",
				DataSource: group.GroupName,
				Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
				Raw:        log,
			}
		}
	}
}
