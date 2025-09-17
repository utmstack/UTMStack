package utils

import (
	"log"
	"os"
)

func Getenv(key string, isMandatory bool) string {
	value, defined := os.LookupEnv(key)
	if !defined {
		if isMandatory {
			log.Fatalf("Error loading environment variable: %s: environment variable does not exist\n", key)
		} else {
			return ""
		}
	}
	if (value == "" || value == " ") && isMandatory {
		log.Fatalf("Error loading environment variable: %s: empty environment variable\n", key)
	}
	return value
}
