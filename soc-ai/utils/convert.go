package utils

import (
	"encoding/json"
	"fmt"
)

func ConvertFromStructToJsonString(alert interface{}) (string, error) {
	bytes, err := json.Marshal(alert)
	if err != nil {
		return "", fmt.Errorf("error marshalling alert: %v", err)
	}

	return string(bytes), nil
}

func ConvertFromJsonToStruct[responseType any](jsonString string) (responseType, error) {
	var response responseType
	err := json.Unmarshal([]byte(jsonString), &response)
	if err != nil {
		return *new(responseType), fmt.Errorf("error unmarshalling GPT response: %v", err)
	}

	return response, nil
}
