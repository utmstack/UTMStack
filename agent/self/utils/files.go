package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// GetMyPath returns the directory path where the currently running executable is located.
// Returns a string representing the directory path, and an error if any error occurs during the process.
func GetMyPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}

// ReadJson reads the json data from the specified file URL and unmarshal it into the provided result interface{}.
// Returns an error if any error occurs during the process.
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

// ReadYAML reads the YAML data from the specified file URL and deserializes it into the provided result interface{}.
// Returns an error if any error occurs during the process.
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

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
