package utils

import (
	"os"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

func Getenv(key string) string {
	value, defined := os.LookupEnv(key)
	if !defined {
		_ = catcher.Error("Error loading environment variable, environment variable does not exist", nil, map[string]any{"key": key})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	if (value == "") || (value == " ") {
		_ = catcher.Error("Error loading environment variable, empty environment variable", nil, map[string]any{"key": key})
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
	return value
}
