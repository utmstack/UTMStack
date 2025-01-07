package processor

import (
	"github.com/threatwinds/go-sdk/catcher"
	"github.com/threatwinds/go-sdk/plugins"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/UTMStack/plugins/o365/configuration"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime string, endTime string, group types.ModuleGroup) {
	agent := GetOfficeProcessor(group)

	err := agent.GetAuth()
	if err != nil {
		_ = catcher.Error("error getting auth", err, map[string]any{})
		return
	}

	err = agent.StartSubscriptions()
	if err != nil {
		_ = catcher.Error("error starting subscriptions", err, map[string]any{})
		return
	}

	logs := agent.GetLogs(startTime, endTime)
	for _, log := range logs {
		LogQueue <- &plugins.Log{
			Id:         uuid.New().String(),
			TenantId:   configuration.GetTenantId(),
			DataType:   "o365",
			DataSource: group.GroupName,
			Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
			Raw:        log,
		}
	}
}
