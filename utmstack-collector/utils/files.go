package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func ReadYAML(path string, result interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	d := yaml.NewDecoder(file)
	if err := d.Decode(result); err != nil {
		return err
	}

	return nil
}

func WriteStringToFile(fileName string, body string) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(body)
	return err
}

func WriteYAML(url string, data interface{}) error {
	config, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	err = WriteStringToFile(url, string(config))
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

	err = WriteStringToFile(path, string(jsonData))
	if err != nil {
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

func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return Logger.ErrorF("error creating path: %v", err)
		}
	} else if err != nil {
		return Logger.ErrorF("error checking path: %v", err)
	}
	return nil
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
