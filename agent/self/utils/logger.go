package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	SelfLogger         *logger.Logger
	loggerOnceInstance sync.Once
)

func InitLogger(filename string) {
	loggerOnceInstance.Do(func() {
		SelfLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 100, Output: filename, Retries: 3, Wait: 5},
		)
	})
}
