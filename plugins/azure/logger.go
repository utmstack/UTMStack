package main

import (
	"github.com/threatwinds/logger"
)

var Logger *logger.Logger

func initLogger(level int) {
	if level == 0 {
		level = 400
	}

	Logger = logger.NewLogger(&logger.Config{
		Format: "text",
		Level:  level,
	})
}
