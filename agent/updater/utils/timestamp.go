package utils

import "time"

// GetCurrentTime returns the current time in the format YYYYMMDDHHMMSS
func GetCurrentTime() string {
	t := time.Now()
	return t.Format("20060102150405")
}
