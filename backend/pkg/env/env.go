package env

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// LoadDotEnv loads variables from a .env file in the current working directory
// into the process environment. Variables already set in the environment take
// precedence (godotenv.Load does not override). A missing .env is not an error.
func LoadDotEnv() {
	if err := godotenv.Load(); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return
		}
		log.Printf("env: failed to load .env: %v", err)
	}
}

func String(key, fallback string, required bool) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	if required {
		log.Fatalf("env: required variable %s is not set", key)
	}
	return fallback
}

func Int(key string, fallback int, required bool) int {
	if v, ok := os.LookupEnv(key); ok {
		n, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("env: variable %s is not a valid int: %v", key, err)
		}
		return n
	}
	if required {
		log.Fatalf("env: required variable %s is not set", key)
	}
	return fallback
}

func Bool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
