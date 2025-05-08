package processor

import (
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/sophos/utils"
	"github.com/utmstack/config-client-go/types"
)

const delayCheck = 300

var timeGroups = make(map[int]int)
var nextKeys = make(map[int]string)

func PullLogs(group types.ModuleGroup) *logger.Error {
	utils.Logger.Info("starting log sync for : %s", group.GroupName)

	epoch := int(time.Now().Unix())

	_, ok := timeGroups[group.ModuleID]
	if !ok {
		timeGroups[group.ModuleID] = epoch - delayCheck
	}

	defer func() {
		timeGroups[group.ModuleID] = epoch + 1
	}()

	agent := getSophosCentralProcessor(group)

	logs, newNextKey, err := agent.getLogs(timeGroups[group.ModuleID], nextKeys[group.ModuleID], group)
	if err != nil {
		return err
	}

	nextKeys[group.ModuleID] = newNextKey

	err = SendToLogstash(logs)
	if err != nil {
		return err
	}

	return nil
}
