package utils

import "time"

func GetCurrentTime() string {
	t := time.Now()
	return t.Format("20060102150405")
}
