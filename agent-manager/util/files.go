package util

import (
	"encoding/json"
	"os"
)

func ReadJson(fileName string, data interface{}) error {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, data)
	if err != nil {
		return err
	}
	return nil
}
