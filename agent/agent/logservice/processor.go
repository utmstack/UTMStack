package logservice

import (
	context "context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/quantfall/holmes"
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
	processor     LogProcessor
	processorOnce sync.Once
	LogQueue      = make(chan LogPipe, 1000)
)

func GetLogProcessor() LogProcessor {
	processorOnce.Do(func() {
		processor = LogProcessor{}
	})
	return processor
}

func (l *LogProcessor) ProcessLogs(client LogServiceClient, ctx context.Context, cnf *configuration.Config, h *holmes.Logger) {
	connectionTime := 0 * time.Second
	reconnectDelay := configuration.InitialReconnectDelay
	invalidKeyCounter := 0

	path, err := utils.GetMyPath()
	if err != nil {
		h.FatalError("Failed to get current path: %v", err)
	}

	filePath := filepath.Join(path, "logs_process")
	utils.CreatePathIfNotExist(filePath)
	fileNames := map[string]*os.File{}
	defer func() {
		for _, file := range fileNames {
			file.Close()
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
			h.Error("Error sending logs to Log Auth Proxy: %v", err)
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					h.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					err := agent.UninstallAll()
					if err != nil {
						h.Error("Error uninstalling agent: %s", err)
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
			h.Error("Error sending logs to Log Auth Proxy: %s", rcv.Message)
			if strings.Contains(rcv.Message, "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					h.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					err := agent.UninstallAll()
					if err != nil {
						h.Error("Error uninstalling agent: %s", err)
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

		invalidKeyCounter = 0

		fileIsOpen := false
		for name := range fileNames {
			if name == filepath.Join(filePath, string(newLog.Src)+".txt") {
				fileIsOpen = true
			}
		}

		newFileName := filepath.Join(filePath, string(newLog.Src)+".txt")
		if !fileIsOpen {
			file, err := os.OpenFile(newFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				h.Error("error opening file: %s", err)
				time.Sleep(reconnectDelay)
				connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
				continue
			}
			fileNames[newFileName] = file
		}

		for _, mylog := range newLog.Logs {
			_, err = fileNames[newFileName].WriteString(fmt.Sprintf("%s\n", mylog))
			if err != nil {
				h.Info("error writing to file: %s\n", err)
				time.Sleep(reconnectDelay)
				connectionTime = utils.IncrementReconnectTime(connectionTime, reconnectDelay, configuration.MaxConnectionTime)
				reconnectDelay = utils.IncrementReconnectDelay(reconnectDelay, configuration.MaxReconnectDelay)
				continue
			}
		}
	}
}

func (l *LogProcessor) ProcessLogsWithHighPriority(msg string, client LogServiceClient, ctx context.Context, cnf *configuration.Config, h *holmes.Logger) error {
	host, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("error getting hostname: %v", err)
	}

	rcv, err := client.ProcessLogs(ctx, &LogMessage{LogType: string(configuration.LogTypeGeneric), Data: []string{"[utm_stack_agent_ds=" + host + "]-" + msg}})
	if err != nil {
		return fmt.Errorf("error sending logs to Log Auth Proxy: %v", err)
	}
	if !rcv.Received {
		return fmt.Errorf("error sending logs to Log Auth Proxy: %s", rcv.Message)
	}
	return nil
}
