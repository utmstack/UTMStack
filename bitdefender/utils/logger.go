package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	bLogger            *logger.Logger
	loggerOnceInstance sync.Once
)

// GetLogger returns a single instance of a Logger configured to save logs to a rotating file.
func GetLogger() *logger.Logger {
	loggerOnceInstance.Do(func() {
		bLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 200, Output: "stdout", Retries: 3, Wait: 5},
		)
	})
	return bLogger
}
