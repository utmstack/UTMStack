package utils

import (
	"fmt"
	"regexp"
	"time"

	"github.com/threatwinds/logger"
)

const maxBatchWait = 5 * time.Second

func TailLogFile(filePath string, logLinesChan chan []string, stopChan chan bool, batchCapacity int, h *logger.Logger) {
	latestline := "null"
	batch := make([]string, 0, batchCapacity)
	ticker := time.NewTicker(maxBatchWait)

loop:
	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				logLinesChan <- batch
				batch = make([]string, 0, batchCapacity)
			}
		case <-stopChan:
			if len(batch) > 0 {
				logLinesChan <- batch
			}
			break loop

		default:
			lines, err := ReadFileLines(filePath)
			if err != nil {
				h.Info("error reading file %s: %v\n", filePath, err)
				continue
			}
			if len(lines) <= 1 {
				if len(lines) == 1 && latestline != lines[0] {
					batch = append(batch, lines[0])
					latestline = lines[0]
				}
			} else if latestline != lines[len(lines)-1] && (len(lines) != 0) {
				if latestline != "null" {
					var index int
					for i, v := range lines {
						if v == latestline {
							index = i
						}
					}
					lines = lines[index+1:]
				}
				for _, line := range lines {
					batch = append(batch, line)
					if len(batch) == cap(batch) {
						logLinesChan <- batch
						batch = make([]string, 0, batchCapacity)
					}
				}
				latestline = lines[len(lines)-1]
			}
			time.Sleep(time.Second)
		}
	}
}

func WatchFolder(logType string, logsPath string, logLinesChan chan []string, batchCapacity int, h *logger.Logger) {
	stopChan := make(chan bool)
	latestLog := ""
	pattern := regexp.MustCompile(fmt.Sprintf(`%s-(\d+)(?:-(\d+))?\.ndjson`, logType))

	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		isEmpty, err := IsDirEmpty(logsPath)
		if err != nil {
			h.Info("error checking if %s is empty: %v\n", logsPath, err)
			continue
		}
		if !isEmpty {
			newLatestLog, err := FindLatestLog(logsPath, pattern)
			if err != nil {
				h.Info("error getting latest log name: %v", err)
				continue
			}
			if newLatestLog != latestLog && newLatestLog != "" {
				if latestLog != "" {
					stopChan <- true
				}
				latestLog = newLatestLog
				go TailLogFile(latestLog, logLinesChan, stopChan, batchCapacity, h)
			}
		}
	}
}
