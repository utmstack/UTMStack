package processor

import (
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/sophos/utils"
	"github.com/utmstack/config-client-go/types"
)

var nextKeys = make(map[int]string)

func PullLogs(group types.ModuleGroup, startTimeStr, endTimeStr string) *logger.Error {
	utils.Logger.Info("starting log sync for : %s from %s to %s", group.GroupName, startTimeStr, endTimeStr)

	startTime, err := time.Parse(time.RFC3339, startTimeStr)
	if err != nil {
		return utils.Logger.ErrorF("error parsing start time: %v", err)
	}

	startEpoch := int(startTime.Unix())

	agent := getSophosCentralProcessor(group)

	logs, newNextKey, logErr := agent.getLogs(startEpoch, nextKeys[group.ModuleID], group)
	if logErr != nil {
		return logErr
	}

	nextKeys[group.ModuleID] = newNextKey

	sendErr := SendToLogstash(logs)
	if sendErr != nil {
		return sendErr
	}

	return nil
}
