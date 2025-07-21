package processor

import (
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/aws/utils"
	"github.com/utmstack/config-client-go/types"
)

func PullLogs(startTime time.Time, endTime time.Time, group types.ModuleGroup) *logger.Error {
	utils.Logger.Info("starting log sync for : %s from %s to %s", group.GroupName, startTime, endTime)

	agent := GetAWSProcessor(group)

	logs, err := agent.GetLogs(startTime, endTime, group)
	if err != nil {
		return err
	}

	err = SendToLogstash(logs)
	if err != nil {
		return err
	}

	return nil
}
