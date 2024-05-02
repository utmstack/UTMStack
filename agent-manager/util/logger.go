package util

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	aLogger            *logger.Logger
	loggerOnceInstance sync.Once
)

func GetLogger() *logger.Logger {
	loggerOnceInstance.Do(func() {
		aLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 200, Output: "stdout"},
		)
	})
	return aLogger
}
