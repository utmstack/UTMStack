package fs

import (
	"fmt"
	"os"
	"reflect"

	"gopkg.in/yaml.v3"
)

// ReadYAML reads a YAML file and unmarshals it into the provided data structure.
// The data parameter must be a non-nil pointer.
func ReadYAML(path string, data interface{}) error {
	if data == nil {
		return fmt.Errorf("data interface is nil")
	}

	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("data must be a non-nil pointer")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(data); err != nil {
		return fmt.Errorf("error decoding YAML: %w", err)
	}
	return nil
}

// WriteYAML marshals the data and writes it to a YAML file.
func WriteYAML(path string, data interface{}) error {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling YAML: %w", err)
	}

	if err = WriteString(path, string(yamlData)); err != nil {
		return fmt.Errorf("error writing YAML file: %w", err)
	}
	return nil
}
