package processor

import (
	"github.com/utmstack/UTMStack/office365/utils"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime string, endTime string, group types.ModuleGroup) {
	utils.Logger.Info("starting log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		return
	}

	agent.GetLogs(startTime, endTime, group)
}
