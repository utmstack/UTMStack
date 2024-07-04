package utils

import (
	"sync"

	"github.com/threatwinds/logger"
)

var (
	Logger             *logger.Logger
	loggerOnceInstance sync.Once
)

func InitLogger(filename string) {
	loggerOnceInstance.Do(func() {
		Logger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 200, Output: filename, Retries: 3, Wait: 5},
		)
	})
}
