package utils

import (
	"github.com/threatwinds/logger"
)

var Logger *logger.Logger

func InitLogger(level int) {
	if level == 0 {
		level = 400
	}

	Logger = logger.NewLogger(&logger.Config{
		Format: "text",
		Level:  level,
	})
}
