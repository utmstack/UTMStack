package processor

import (
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

var timeGroups map[int]int

func PullLogs(group types.ModuleGroup) *logger.Error {
	epoch := time.Now().Second()

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
		return err
	}

	err = SendToCorrelation(logs)
	if err != nil {
		return err
	}

	return nil
}
