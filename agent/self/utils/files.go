package utils

import (
	"os"
	"path/filepath"
)

func GetMyPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
}
