package utils

import (
	"os"

	"gopkg.in/yaml.v2"
)

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

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
