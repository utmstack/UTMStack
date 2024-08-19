package processor

import (
	"encoding/json"
	"fmt"
)

type TransformedLog struct {
	Logx struct {
		Google map[string]interface{} `json:"google"`
	} `json:"logx"`
}

func ETLProcess(jsonStr string) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	transformedLog := &TransformedLog{}
	transformedLog.Logx.Google = data

	jsonData, err := json.Marshal(transformedLog)
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}
