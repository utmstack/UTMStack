package processor

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func etlProcess(events []*cloudwatchlogs.OutputLogEvent) []string {
	var logs = []string{}

	for _, event := range events {
		logs = append(logs, *event.Message)
	}

	return logs
}
