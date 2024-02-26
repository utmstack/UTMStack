package utils

import (
	"log"
	"os"
	"strconv"

	"github.com/threatwinds/logger"
)

var Logger *logger.Logger

func init() {
	lenv := os.Getenv("LOG_LEVEL")
	var level int
	var err error

	if lenv != "" && lenv != " " {
		level, err = strconv.Atoi(lenv)
		if err != nil {
			log.Fatalln(err)
		}
	} else {
		level = 400
	}

	Logger = logger.NewLogger(&logger.Config{
		Format: "text",
		Level:  level,
	})
}