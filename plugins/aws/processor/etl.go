package processor

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

type TransformedLog struct {
	Logx struct {
		AWS map[string]interface{} `json:"aws"`
	} `json:"logx"`
}

func ETLProcess(events []*cloudwatchlogs.OutputLogEvent, logGroup, logStream string) ([]string, error) {
	var logs = []string{}

	for _, event := range events {
		transformedLog := TransformedLog{}

		transformedLog.Logx.AWS = ProcessAwsFlowLog(*event.Message)
		transformedLog.Logx.AWS["group"] = logGroup
		transformedLog.Logx.AWS["stream"] = logStream

		jsonData, err := json.Marshal(transformedLog)
		if err != nil {
			return nil, err
		}

		logs = append(logs, string(jsonData))
	}

	return logs, nil
}
