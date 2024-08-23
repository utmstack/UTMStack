package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetMyPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}

func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return fmt.Errorf("error creating path: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("error checking path: %v", err)
	}
	return nil
}
