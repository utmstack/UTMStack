package utils

import (
	"fmt"
	"regexp"
	"time"
)

func TailLogFile(filePath string, logLinesChan chan string, stopChan chan bool) {
	latestline := "null"

loop:
	for {
		select {
		case <-stopChan:
			break loop

		default:
			lines, err := ReadFileLines(filePath)
			if err != nil {
				Logger.Info("error reading file %s: %v\n", filePath, err)
				continue
			}
			if len(lines) == 1 && latestline != lines[0] {
				logLinesChan <- lines[0]
				latestline = lines[0]
			} else if len(lines) > 1 && latestline != lines[len(lines)-1] {
				var startIndex int = 0
				if latestline != "null" {
					for i, v := range lines {
						if v == latestline {
							startIndex = i + 1
							break
						}
					}
				}
				for _, line := range lines[startIndex:] {
					logLinesChan <- line
				}
				if len(lines) > 0 {
					latestline = lines[len(lines)-1]
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func WatchFolder(logType string, logsPath string, logLinesChan chan string) {
	stopChan := make(chan bool)
	latestLog := ""
	pattern := regexp.MustCompile(fmt.Sprintf(`%s-(\d+)(?:-(\d+))?\.ndjson`, logType))

	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		isEmpty, err := IsDirEmpty(logsPath)
		if err != nil {
			Logger.Info("error checking if %s is empty: %v\n", logsPath, err)
			continue
		}
		if !isEmpty {
			newLatestLog, err := FindLatestLog(logsPath, pattern)
			if err != nil {
				Logger.Info("error getting latest log name: %v", err)
				continue
			}
			if newLatestLog != latestLog && newLatestLog != "" {
				if latestLog != "" {
					stopChan <- true
				}
				latestLog = newLatestLog
				go TailLogFile(latestLog, logLinesChan, stopChan)
			}
		}
	}
}
