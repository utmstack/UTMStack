package utils

import (
	"os"
	"path/filepath"

	"github.com/threatwinds/go-sdk/catcher"
)

func GetMyPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", err
	}
	exPath := filepath.Dir(ex)
	return exPath, nil
}

// CreatePathIfNotExist creates a specific path if not exist
func CreatePathIfNotExist(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			return catcher.Error("error creating path", err, map[string]any{"path": path})
		}
	} else if err != nil {
		return catcher.Error("error checking path", err, map[string]any{"path": path})
	}
	return nil
}

func CheckIfPathExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
