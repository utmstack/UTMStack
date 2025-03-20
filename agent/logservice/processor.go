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

	"github.com/utmstack/UTMStack/agent/agent"
	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/conn"
	"github.com/utmstack/UTMStack/agent/utils"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type LogProcessor struct {
	connErrWritten bool
	sendErrWritten bool
}

type LogPipe struct {
	Src  string
	Logs []string
}

var (
	processor                   LogProcessor
	processorOnce               sync.Once
	LogQueue                    = make(chan LogPipe, 1000)
	MinutesForCleanLog          = 10080
	MinutesForReportLogsCounted = time.Duration(5 * time.Minute)
	timeToSleep                 = 10 * time.Second
	logsProcessCounter          = map[string]int{}
)

func GetLogProcessor() LogProcessor {
	processorOnce.Do(func() {
		processor = LogProcessor{
			connErrWritten: false,
			sendErrWritten: false,
		}
	})
	return processor
}

func (l *LogProcessor) ProcessLogs(cnf *config.Config, ctx context.Context) {
	go func() {
		for {
			time.Sleep(MinutesForReportLogsCounted)
			SaveCountedLogs()
			logsProcessCounter = map[string]int{}
		}
	}()

	for {
		ctxEof, cancelEof := context.WithCancel(context.Background())
		connection, err := conn.GetCorrelationConnection(cnf)
		if err != nil {
			if !l.connErrWritten {
				utils.Logger.ErrorF("error connecting to Correlation: %v", err)
				l.connErrWritten = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		client := NewLogServiceClient(connection)
		l.connErrWritten = false
		l.processLogs(client, ctx, ctxEof, cancelEof)
	}
}

func (l *LogProcessor) processLogs(client LogServiceClient, ctx context.Context, ctxEof context.Context, cancel context.CancelFunc) {
	invalidKeyCounter := 0

	for {
		select {
		case <-ctxEof.Done():
			utils.Logger.Info("LogProcessor: Context done, exiting...")
			return
		case newLog := <-LogQueue:
			rcv, err := client.ProcessLogs(ctx, &LogMessage{Type: agent.ConnectorType_AGENT, LogType: newLog.Src, Data: newLog.Logs})
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					time.Sleep(timeToSleep)
					cancel()
					return
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !l.sendErrWritten {
						utils.Logger.ErrorF("failed to send log: %v", err)
						l.sendErrWritten = true
					}
					time.Sleep(timeToSleep)
					cancel()
					return
				} else {
					if !l.sendErrWritten {
						utils.Logger.ErrorF("failed to send log: %v ", err)
						l.sendErrWritten = true
					}
					time.Sleep(timeToSleep)
					continue
				}
			} else if !rcv.Received {
				utils.Logger.ErrorF("Error sending logs to Log Auth Proxy: %s", rcv.Message)
				if strings.Contains(rcv.Message, "invalid agent key") {
					invalidKeyCounter++
					if invalidKeyCounter >= 20 {
						utils.Logger.Info("Uninstalling agent: reason: agent has been removed from the panel...")
						err := agent.UninstallAll()
						if err != nil {
							utils.Logger.ErrorF("Error uninstalling agent: %s", err)
						}
					}
				} else {
					invalidKeyCounter = 0
				}

				time.Sleep(timeToSleep)
				continue
			}

			l.sendErrWritten = false
			logsProcessCounter[newLog.Src] += len(newLog.Logs)
			invalidKeyCounter = 0
		}

	}
}

func SaveCountedLogs() {
	path := utils.GetMyPath()
	filePath := filepath.Join(path, "logs_process")
	logFile := filepath.Join(filePath, "processed_logs.txt")
	utils.CreatePathIfNotExist(filePath)

	file, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		utils.Logger.ErrorF("error opening processed_logs.txt file: %s", err)
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
			utils.Logger.ErrorF("error parsing first log time: %s", err)
			return
		}

		if !firstLogTime.IsZero() && time.Since(firstLogTime).Minutes() >= float64(MinutesForCleanLog) {
			file.Close()
			if err := os.Remove(logFile); err != nil {
				utils.Logger.ErrorF("error removing processed_logs.txt file: %s", err)
				return
			}
			file, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
			if err != nil {
				utils.Logger.ErrorF("error opening processed_logs.txt file: %s", err)
				return
			}
		}
	}

	for name, counter := range logsProcessCounter {
		if counter > 0 {
			_, err = file.WriteString(fmt.Sprintf("%v - %d logs from %s have been processed\n", time.Now().Format("2006/01/02 15:04:05.9999999 -0700 MST"), counter, name))
			if err != nil {
				utils.Logger.ErrorF("error writing to processed_logs.txt file: %s", err)
				continue
			}
		}
	}

}
