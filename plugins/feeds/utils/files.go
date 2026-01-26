package utils

import (
	"os"
	"reflect"

	"github.com/threatwinds/go-sdk/catcher"
	"gopkg.in/yaml.v2"
)

func ReadYAML(path string, result interface{}) error {
	if result == nil {
		return catcher.Error("result interface is nil", nil, nil)
	}

	rv := reflect.ValueOf(result)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return catcher.Error("result must be a non-nil pointer", nil, nil)
	}

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
