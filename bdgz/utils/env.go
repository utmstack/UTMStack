package utils

import (
	"log"
	"os"
)

// Getenv returns the environment variable
func Getenv(key string) string {
	value, defined := os.LookupEnv(key)
	if !defined {
		log.Fatalf("Error loading environment variable: %s: environment variable does not exist\n", key)
	}
	if (value == "") || (value == " ") {
		log.Fatalf("Error loading environment variable: %s: empty environment variable\n", key)
	}
	return value
}
