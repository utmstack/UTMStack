package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	selfLogger         *logger.Logger
	loggerOnceInstance sync.Once
)

func CreateLogger(filename string) *logger.Logger {
	loggerOnceInstance.Do(func() {
		selfLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 100, Output: filename, Retries: 3, Wait: 5},
		)
	})
	return selfLogger
}
