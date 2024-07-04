package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	ALogger            *logger.Logger
	loggerOnceInstance sync.Once
)

func InitLogger() {
	loggerOnceInstance.Do(func() {
		ALogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 200, Output: "stdout"},
		)
	})
}
