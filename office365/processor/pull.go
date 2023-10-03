package processor

import (
	"github.com/quantfall/holmes"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime string, endTime string, group types.ModuleGroup, h *holmes.Logger) {
	h.Info("Init log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetOfficeProcessor(group)
	err := agent.GetAuth()
	if err != nil {
		h.Error("failed to authenticate with Office365 API: %v", err)
		return
	}
	err = agent.StartSubscriptions(h)
	if err != nil {
		h.Error("failed to configure subscriptions: %v", err)
		return
	}
	err = agent.GetLogs(startTime, endTime, group, h)
	if err != nil {
		h.Error("failed to get new logs: %v", err)
		return
	}
}
