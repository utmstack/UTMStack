package util

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

func ReadYAML(path string, result interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	d := yaml.NewDecoder(file)
	if err := d.Decode(result); err != nil {
		return err
	}
	return nil
}

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

func WriteJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	err = writeToFile(path, string(jsonData[:]))
	if err != nil {
		return err
	}

	return nil
}

func writeToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)

	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.WriteString(body)
	return err
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("error creating path: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %v", err)
	}
	return nil
}
