package main

import "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"

func etlProcess(events []types.OutputLogEvent) []string {
	var logs = make([]string, 0)

	for _, event := range events {
		logs = append(logs, *event.Message)
	}

	return logs
}
