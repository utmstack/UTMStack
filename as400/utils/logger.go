package utils

import (
	"path/filepath"
	"sync"

	"github.com/threatwinds/logger"
)

var (
	Logger             *logger.Logger
	loggerOnceInstance sync.Once
	logLevelConfigFile = filepath.Join(GetMyPath(), "log_level.yml")
	LogLevelMap        = map[string]int{
		"debug":    100,
		"info":     200,
		"notice":   300,
		"warning":  400,
		"error":    500,
		"critical": 502,
		"alert":    509,
	}
)

type LogLevels struct {
	Level string `yaml:"level"`
}

func InitLogger(filename string) {
	logLevel := LogLevels{}
	err := ReadYAML(logLevelConfigFile, &logLevel)
	if err != nil {
		logLevel.Level = "info"
	}
	logLevelInt := 200
	if val, ok := LogLevelMap[logLevel.Level]; ok {
		logLevelInt = val
	}
	loggerOnceInstance.Do(func() {
		Logger = logger.NewLogger(
			&logger.Config{Format: "text", Level: logLevelInt, Output: filename, Retries: 3, Wait: 5},
		)
	})
}
