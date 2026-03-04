package utils

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"
)

func TailLogFile(filePath string, logLinesChan chan string, stopChan chan struct{}) {
	var offset int64

	// Start from end of file to avoid re-sending old lines
	if info, err := os.Stat(filePath); err == nil {
		offset = info.Size()
	}

	for {
		select {
		case <-stopChan:
			return
		default:
			newOffset, err := readNewLines(filePath, offset, logLinesChan)
			if err != nil {
				Logger.Info("error reading file %s: %v", filePath, err)
			} else {
				offset = newOffset
			}
			time.Sleep(time.Second)
		}
	}
}

func readNewLines(filePath string, offset int64, logLinesChan chan string) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return offset, err
	}
	defer file.Close()

	// Check if file was truncated or rotated
	info, err := file.Stat()
	if err != nil {
		return offset, err
	}
	if info.Size() < offset {
		offset = 0
	}

	if info.Size() == offset {
		return offset, nil
	}

	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return offset, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "" {
			logLinesChan <- line
		}
	}

	newOffset, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return offset, err
	}
	return newOffset, scanner.Err()
}

func WatchFolder(logType string, logsPath string, logLinesChan chan string) {
	var currentStopChan chan struct{}
	latestLog := ""
	pattern := regexp.MustCompile(fmt.Sprintf(`%s-(\d+)(?:-(\d+))?\.ndjson`, logType))

	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		isEmpty, err := IsDirEmpty(logsPath)
		if err != nil {
			Logger.Info("error checking if %s is empty: %v", logsPath, err)
			continue
		}
		if !isEmpty {
			newLatestLog, err := FindLatestLog(logsPath, pattern)
			if err != nil {
				Logger.Info("error getting latest log name: %v", err)
				continue
			}
			if newLatestLog != latestLog && newLatestLog != "" {
				if currentStopChan != nil {
					close(currentStopChan)
				}
				latestLog = newLatestLog
				currentStopChan = make(chan struct{})
				go TailLogFile(latestLog, logLinesChan, currentStopChan)
			}
		}
	}
}
