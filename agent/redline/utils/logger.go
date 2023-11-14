package utils

import (
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	logger             *lumberjack.Logger
	loggerOnceInstance sync.Once
)

// CreateLogger returns a single instance of a Logger configured to save logs to a rotating file.
func CreateLogger(filename string) *lumberjack.Logger {
	loggerOnceInstance.Do(func() {
		logger = &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    5,
			MaxBackups: 100,
			MaxAge:     30,
		}
	})
	return logger
}
