package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	UpdaterLogger      *logger.Logger
	loggerOnceInstance sync.Once
)

func InitLogger(filename string) {
	loggerOnceInstance.Do(func() {
		UpdaterLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 100, Output: filename, Retries: 3, Wait: 5},
		)
	})
}
