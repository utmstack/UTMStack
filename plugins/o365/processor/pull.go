package processor

import (
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/plugins/o365/configuration"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime string, endTime string, group types.ModuleGroup) {
	go_sdk.Logger().Info("starting log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		go_sdk.Logger().ErrorF("error getting auth token: %v", err)
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		go_sdk.Logger().ErrorF("error starting subscriptions: %v", err)
		return
	}

	logs := agent.GetLogs(startTime, endTime, group)
	for _, log := range logs {
		LogQueue <- &go_sdk.Log{
			Id:         uuid.New().String(),
			TenantId:   configuration.GetTenantId(),
			DataType:   "o365",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		}
	}
}
