package fs

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadJSON reads a JSON file and unmarshals it into the provided data structure.
func ReadJSON(path string, data interface{}) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	if err = json.Unmarshal(content, data); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}
	return nil
}

// WriteJSON marshals the data and writes it to a JSON file with indentation.
func WriteJSON(path string, data interface{}) error {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if err = WriteString(path, string(jsonData)); err != nil {
		return fmt.Errorf("error writing JSON file: %w", err)
	}
	return nil
}
