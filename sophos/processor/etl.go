package processor

import (
	"time"

	"github.com/utmstack/config-client-go/types"
)

type TransformedLog struct {
	Logx struct {
		SophosCentral map[string]interface{} `json:"sophos_central"`
	} `json:"logx"`
	Global struct {
		Type string `json:"type"`
	} `json:"global"`
	Timestamp  string `json:"@timestamp"`
	DataType   string `json:"dataType"`
	DataSource string `json:"dataSource"`
}

func ETLProcess(events EventAggregate, group types.ModuleGroup) []TransformedLog {
	var logs []TransformedLog

	for _, event := range events.Items {
		transformedLog := TransformedLog{}
		transformedLog.Timestamp = time.Now().Format("2006-01-02T15:04:05.000Z")

		transformedLog.DataType = "sophos-central"
		transformedLog.DataSource = group.GroupName
		transformedLog.Global.Type = "logx"
		transformedLog.Logx.SophosCentral = event

		logs = append(logs, transformedLog)
	}

	return logs
}
