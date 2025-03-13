package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

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
