package config

import (
	"sync"

	twlogger "github.com/threatwinds/logger"
)

var (
	iLogger            *twlogger.Logger
	loggerOnceInstance sync.Once
)

func Logger() *twlogger.Logger {
	loggerOnceInstance.Do(func() {
		iLogger = twlogger.NewLogger(
			&twlogger.Config{Format: "text", Level: 200, Output: ServiceLogPath, Retries: 3, Wait: 5},
		)
	})
	return iLogger
}
