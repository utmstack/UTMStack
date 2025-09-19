package processor

import (
	"time"

	"github.com/threatwinds/go-sdk/catcher"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime time.Time, endTime time.Time, group types.ModuleGroup) error {
	catcher.Info("starting log sync", map[string]any{"group": group.GroupName, "start_time": startTime, "end_time": endTime})

	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime, group)
	if err != nil {
		return catcher.Error("error pulling logs", err, map[string]any{"group": group.GroupName})
	}

	err = SendToLogstash(logs)
	if err != nil {
		return catcher.Error("error sending logs to logstash", err, map[string]any{"group": group.GroupName})
	}

	return nil
}
