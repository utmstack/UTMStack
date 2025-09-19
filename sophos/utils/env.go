package utils

import (
	"os"

	"github.com/threatwinds/go-sdk/catcher"
)

// Getenv returns the environment variable
func Getenv(key string) string {
	value, defined := os.LookupEnv(key)
	if !defined {
		catcher.Error("Error loading environment variable, environment variable does not exist", nil, map[string]any{"key": key})
		os.Exit(1)
	}
	if (value == "") || (value == " ") {
		catcher.Error("Error loading environment variable, empty environment variable", nil, map[string]any{"key": key})
		os.Exit(1)
	}
	return value
}
