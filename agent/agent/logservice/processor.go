package logservice

import (
	"bufio"
	context "context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/logger"
	"github.com/utmstack/UTMStack/agent/agent/agent"
	"github.com/utmstack/UTMStack/agent/agent/configuration"
	"github.com/utmstack/UTMStack/agent/agent/utils"
)

type LogProcessor struct {
}

type LogPipe struct {
	Src  string
	Logs []string
}

var (
	processor                   LogProcessor
	processorOnce               sync.Once
	LogQueue                    = make(chan LogPipe, 1000)
	MinutesForCleanLog          = 10080 // 7 days in minutes(7*24*60)
	MinutesForReportLogsCounted = time.Duration(5 * time.Minute)
)

func GetLogProcessor() LogProcessor {
	processorOnce.Do(func() {
		processor = LogProcessor{}
	})
	return processor
}

func (l *LogProcessor) ProcessLogs(client LogServiceClient, ctx context.Context, cnf *configuration.Config, h *logger.Logger) {
	connectionTime := 0 * time.Second
	reconnectDelay := configuration.InitialReconnectDelay
	invalidKeyCounter := 0

	logsProcessCounter := map[string]int{}
	go func() {
		for {
			time.Sleep(MinutesForReportLogsCounted)
			SaveCountedLogs(h, logsProcessCounter)
			logsProcessCounter = map[string]int{}
		}
	}()

	for {
		if connectionTime >= configuration.MaxConnectionTime {
			connectionTime = 0 * time.Second
			reconnectDelay = configuration.InitialReconnectDelay
			continue
		}

		newLog := <-LogQueue
		rcv, err := client.ProcessLogs(ctx, &LogMessage{LogType: newLog.Src, Data: newLog.Logs})
		if err != nil {
			h.ErrorF("Error sending logs to Log Auth Proxy: %v", err)
			for _, log := range newLog.Logs {
				h.ErrorF("log with errors: %s", log)
			}
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					h.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					err := agent.UninstallAll()
					if err != nil {
						h.ErrorF("Error uninstalling agent: %s", err)
					}
				}
			} else {
				invalidKeyCounter = 0
			}

			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
			continue
		} else if !rcv.Received {
			h.ErrorF("Error sending logs to Log Auth Proxy: %s", rcv.Message)
			h.Info("logs with errors: ")
			for _, log := range newLog.Logs {
				h.Info("log: %s", log)
			}
			if strings.Contains(rcv.Message, "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					h.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					err := agent.UninstallAll()
					if err != nil {
						h.ErrorF("Error uninstalling agent: %s", err)
					}
				}
			} else {
				invalidKeyCounter = 0
			}

			time.Sleep(reconnectDelay)
			connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
			reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
			continue
		}

		logsProcessCounter[newLog.Src] += len(newLog.Logs)
		invalidKeyCounter = 0
	}
}

func (l *LogProcessor) ProcessLogsWithHighPriority(msg string, client LogServiceClient, ctx context.Context, cnf *configuration.Config, h *logger.Logger) error {
	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %v", err)
	}

	rcv, err := client.ProcessLogs(ctx, &LogMessage{LogType: string(configuration.LogTypeGeneric), Data: []string{configuration.GetMessageFormated(host, msg)}})
	if err != nil {
		return fmt.Errorf("error sending logs to Log Auth Proxy: %v", err)
	}
	if !rcv.Received {
		return fmt.Errorf("error sending logs to Log Auth Proxy: %s", rcv.Message)
	}
	return nil
}

func SaveCountedLogs(h *logger.Logger, logsProcessCounter map[string]int) {
	path, err := utils.GetMyPath()
	if err != nil {
		h.Fatal("Failed to get current path: %v", err)
	}

	filePath := filepath.Join(path, "logs_process")
	logFile := filepath.Join(filePath, "processed_logs.txt")
	utils.CreatePathIfNotExist(filePath)

	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		h.ErrorF("error opening processed_logs.txt file: %s", err)
		return
	}
	defer file.Close()

	var firstLogTime time.Time
	var firstLine string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		firstLine = scanner.Text()
		break
	}

	if firstLine != "" {
		firstLogTime, err = time.Parse("2006/01/02 15:04:05.9999999 -0700 MST", strings.Split(firstLine, " - ")[0])
		if err != nil {
			h.ErrorF("error parsing first log time: %s", err)
			return
		}

		if !firstLogTime.IsZero() && time.Since(firstLogTime).Minutes() >= float64(MinutesForCleanLog) {
			file.Close()
			if err := os.Remove(logFile); err != nil {
				h.ErrorF("error removing processed_logs.txt file: %s", err)
				return
			}
			file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				h.ErrorF("error opening processed_logs.txt file: %s", err)
				return
			}
		}
	}

	for name, counter := range logsProcessCounter {
		if counter > 0 {
			_, err = file.WriteString(fmt.Sprintf("%v - %d logs from %s have been processed\n", time.Now().Format("2006/01/02 15:04:05.9999999 -0700 MST"), counter, name))
			if err != nil {
				h.ErrorF("error writing to processed_logs.txt file: %s", err)
				continue
			}
		}
	}

}
