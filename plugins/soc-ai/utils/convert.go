package utils

import (
	"encoding/json"

	"github.com/threatwinds/go-sdk/catcher"
)

func ConvertFromStructToJsonString(alert interface{}) (string, error) {
	bytes, err := json.Marshal(alert)
	if err != nil {
		return "", catcher.Error("error marshalling alert", err, nil)
	}

	return string(bytes), nil
}

func ConvertFromJsonToStruct[responseType any](jsonString string) (responseType, error) {
	var response responseType
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return *new(responseType), catcher.Error("error unmarshalling GPT response", err, nil)
	}

	return response, nil
}
