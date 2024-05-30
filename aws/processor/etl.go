package processor

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/utmstack/config-client-go/types"
)

type TransformedLog struct {
	Logx struct {
		Type   string                 `json:"type"`
		Tenant string                 `json:"tenant"`
		AWS    map[string]interface{} `json:"aws"`
	} `json:"logx"`
	Global struct {
		Type     string `json:"type"`
		Analysed int    `json:"analysed"`
	} `json:"global"`
	Timestamp  string `json:"@timestamp"`
	DataType   string `json:"dataType"`
	DataSource string `json:"dataSource"`
}

func ETLProcess(events []*cloudwatchlogs.OutputLogEvent, group types.ModuleGroup, logGroup, logStream string) []TransformedLog {
	var logs []TransformedLog

	for _, event := range events {
		transformedLog := TransformedLog{}
		transformedLog.Timestamp = convertMillisecondsToTimeString(event.Timestamp)

		transformedLog.DataType = "aws"
		transformedLog.DataSource = group.GroupName
		transformedLog.Global.Type = "logx"
		transformedLog.Global.Analysed = 1
		transformedLog.Logx.Type = "aws"
		transformedLog.Logx.Tenant = group.GroupName

		transformedLog.Logx.AWS = ProcessAwsFlowLog(*event.Message)
		transformedLog.Logx.AWS["group"] = logGroup
		transformedLog.Logx.AWS["stream"] = logStream

		logs = append(logs, transformedLog)
	}

	return logs
}

func convertMillisecondsToTimeString(ms *int64) string {
	if ms == nil {
		return ""
	}

	sec := *ms / 1000
	nsec := (*ms % 1000) * 1000000

	t := time.Unix(sec, nsec).UTC()

	return t.Format("2006-01-02T15:04:05.000Z")
}
