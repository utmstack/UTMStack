package processor

import "github.com/utmstack/config-client-go/types"

type TransformedLog struct {
	Logx struct {
		Tenant string                 `json:"tenant"`
		O365   map[string]interface{} `json:"o365"`
	} `json:"logx"`
	Global struct {
		Type     string `json:"type"`
		Analysed int    `json:"analysed"`
	} `json:"global"`
	Timestamp  string `json:"@timestamp"`
	DataType   string `json:"dataType"`
	DataSource string `json:"dataSource"`
}

func ETLProcess(data []map[string]interface{}, group types.ModuleGroup) []TransformedLog {
	var logs []TransformedLog
	for _, log := range data {
		transformedLog := TransformedLog{}
		if creationTime, ok := log["CreationTime"].(string); ok {
			transformedLog.Timestamp = creationTime + ".000Z"
		}
		transformedLog.DataType = "o365"
		transformedLog.DataSource = group.GroupName
		transformedLog.Global.Type = "logx"
		transformedLog.Global.Analysed = 1
		transformedLog.Logx.Tenant = group.GroupName
		delete(log, "CreationTime")
		transformedLog.Logx.O365 = log
		logs = append(logs, transformedLog)
	}
	return logs
}
