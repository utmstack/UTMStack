package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"time"
)

func SentLog(msg string) error {
	// Get current path
	path, err := GetMyPath()
	if err != nil {
		return fmt.Errorf("failed to get current path: %v", err)
	}
	switch runtime.GOOS {
	case "windows":
		err := Execute("eventcreate", path, "/T", "INFORMATION", "/ID", "1000", "/SO", "UTMStackAgent", "/D", msg)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	case "linux":
		err := Execute("logger", path, "-p", "syslog.info", msg)
		if err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}

func FindLatestLog(path string, pattern *regexp.Regexp) (string, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return "", err
	}
	var maxTime time.Time
	maxIndex := 0
	maxFile := ""
	for _, file := range files {
		name := file.Name()
		matches := pattern.FindStringSubmatch(name)
		if matches != nil && !file.IsDir() {
			datePart := matches[1]
			timePart, _ := time.Parse("20060102", datePart)
			index := 0
			if len(matches) > 2 && matches[2] != "" {
				index, _ = strconv.Atoi(matches[2])
			}
			if timePart.After(maxTime) || (timePart.Equal(maxTime) && index > maxIndex) {
				maxTime = timePart
				maxIndex = index
				maxFile = file.Name()
			}
		}
	}
	if maxFile == "" {
		return "", nil
	}
	return filepath.Join(path, maxFile), nil
}
