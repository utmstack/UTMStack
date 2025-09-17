package utils

import (
	"os"
	"strconv"
	"sync"

	"github.com/threatwinds/logger"
)

var (
	ALogger            *logger.Logger
	loggerOnceInstance sync.Once
)

func InitLogger() {
	logLevel := os.Getenv("LOG_LEVEL")
	logLevelInt := 200
	if logLevel != "" {
		logL, err := strconv.Atoi(logLevel)
		if err == nil {
			logLevelInt = logL
		}
	}

	loggerOnceInstance.Do(func() {
		ALogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: logLevelInt, Output: "stdout"},
		)
	})
}
