package processor

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func ETLProcess(events []*cloudwatchlogs.OutputLogEvent) ([]string, error) {
	var logs = []string{}

	for _, event := range events {
		jsonData, err := json.Marshal(ProcessAwsFlowLog(*event.Message))
		if err != nil {
			return nil, err
		}

		logs = append(logs, string(jsonData))
	}

	return logs, nil
}
