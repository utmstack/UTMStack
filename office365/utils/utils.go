package utils

import "github.com/threatwinds/logger"

var Logger *logger.Logger

func init() {
	Logger = logger.NewLogger(&logger.Config{
		Format: "text",
		Level:  200,
	})
}
