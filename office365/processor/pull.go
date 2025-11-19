package processor

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime time.Time, endTime time.Time, group types.ModuleGroup) {
	catcher.Info("starting log sync", map[string]any{"group_name": group.GroupName, "start_time": startTime, "end_time": endTime})

	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		catcher.Error("error getting auth token", err, nil)
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		catcher.Error("error starting subscriptions", err, nil)
		return
	}

	agent.GetLogs(startTime, endTime, group)
}
