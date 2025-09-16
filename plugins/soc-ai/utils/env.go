package utils

import (
	"os"

	"github.com/threatwinds/go-sdk/catcher"
)

func Getenv(key string, isMandatory bool) string {
	value, defined := os.LookupEnv(key)
	if !defined {
		if isMandatory {
			catcher.Error("Error loading environment variable", nil, map[string]any{"key": key})
		} else {
			return ""
		}
	}
	if (value == "" || value == " ") && isMandatory {
		catcher.Error("Error loading environment variable", nil, map[string]any{"key": key})
	}
	return value
}
