package logservice

import (
	context "context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	go_sdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/UTMStack/agent/service/agent"
	"github.com/utmstack/UTMStack/agent/service/config"
	"github.com/utmstack/UTMStack/agent/service/conn"
	"github.com/utmstack/UTMStack/agent/service/database"
	"github.com/utmstack/UTMStack/agent/service/models"
	"github.com/utmstack/UTMStack/agent/service/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogProcessor struct {
	db             *database.Database
	connErrWritten bool
	ackErrWritten  bool
	sendErrWritten bool
}

var (
	processor     LogProcessor
	processorOnce sync.Once
	LogQueue      = make(chan *go_sdk.Log)
	timeToSleep   = time.Duration(10 * time.Second)
	timeCLeanLogs = time.Duration(10 * time.Minute)
)

func GetLogProcessor() LogProcessor {
	processorOnce.Do(func() {
		processor = LogProcessor{
			db:             database.GetDB(),
			connErrWritten: false,
			ackErrWritten:  false,
			sendErrWritten: false,
		}
	})
	return processor
}

func (l *LogProcessor) ProcessLogs(cnf *config.Config, ctx context.Context) {
	go l.CleanCountedLogs()

	for {
		ctxEof, cancelEof := context.WithCancel(context.Background())
		conn, err := conn.GetCorrelationConnection(cnf)
		if err != nil {
			if !l.connErrWritten {
				utils.Logger.ErrorF("error connecting to Correlation: %v", err)
				l.connErrWritten = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		client := go_sdk.NewIntegrationClient(conn)
		plClient := createClient(client, ctx)
		l.connErrWritten = false

		go l.handleAcknowledgements(plClient, ctxEof, cancelEof)
		l.processLogs(plClient, ctxEof, cancelEof)
	}
}

func (l *LogProcessor) handleAcknowledgements(plClient go_sdk.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			ack, err := plClient.Recv()
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					time.Sleep(timeToSleep)
					cancel()
					return
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !l.ackErrWritten {
						utils.Logger.ErrorF("failed to receive ack: %v", err)
						l.ackErrWritten = true
					}
					time.Sleep(timeToSleep)
					cancel()
					return
				} else {
					if !l.ackErrWritten {
						utils.Logger.ErrorF("failed to receive ack: %v", err)
						l.ackErrWritten = true
					}
					time.Sleep(timeToSleep)
					continue
				}
			}

			l.ackErrWritten = false

			l.db.Lock()
			err = l.db.Update(&models.Log{}, "id", ack.LastId, "processed", true)
			if err != nil {
				utils.Logger.ErrorF("failed to update log: %v", err)
			}
			l.db.Unlock()
		}
	}
}

func (l *LogProcessor) processLogs(plClient go_sdk.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("context done, exiting processLogs")
			return
		case newLog := <-LogQueue:
			uuid, err := uuid.NewRandom()
			if err != nil {
				utils.Logger.ErrorF("failed to generate uuid: %v", err)
				continue
			}

			newLog.Id = uuid.String()
			l.db.Lock()
			err = l.db.Create(&models.Log{ID: newLog.Id, Log: newLog.Raw, Type: newLog.DataType, CreatedAt: time.Now(), DataSource: newLog.DataSource, Processed: false})
			if err != nil {
				utils.Logger.ErrorF("failed to save log: %v :log: %s", err, newLog.Raw)
			}
			l.db.Unlock()

			err = plClient.Send(newLog)
			if err != nil {
				if strings.Contains(err.Error(), "EOF") {
					time.Sleep(timeToSleep)
					cancel()
					return
				}
				st, ok := status.FromError(err)
				if ok && (st.Code() == codes.Unavailable || st.Code() == codes.Canceled) {
					if !l.sendErrWritten {
						utils.Logger.ErrorF("failed to send log: %v :log: %s", err, newLog.Raw)
						l.sendErrWritten = true
					}
					time.Sleep(timeToSleep)
					cancel()
					return
				} else {
					if !l.sendErrWritten {
						utils.Logger.ErrorF("failed to send log: %v :log: %s", err, newLog.Raw)
						l.sendErrWritten = true
					}
					time.Sleep(timeToSleep)
					continue
				}
			}
			l.sendErrWritten = false
		}
	}
}

func (l *LogProcessor) CleanCountedLogs() {
	ticker := time.NewTicker(timeCLeanLogs)
	defer ticker.Stop()
	for range ticker.C {
		dataRetention, err := GetDataRetention()
		if err != nil {
			utils.Logger.ErrorF("error getting data retention: %s", err)
			continue
		}
		l.db.Lock()
		_, err = l.db.DeleteOld(&models.Log{}, dataRetention)
		if err != nil {
			utils.Logger.ErrorF("error deleting old logs: %s", err)
		}
		l.db.Unlock()

		unprocessed := []models.Log{}
		l.db.Lock()
		found, err := l.db.Find(&unprocessed, "processed", false)
		l.db.Unlock()
		if err != nil {
			utils.Logger.ErrorF("error finding unprocessed logs: %s", err)
			continue
		}

		if found {
			for _, log := range unprocessed {
				LogQueue <- &go_sdk.Log{
					Id:         log.ID,
					Raw:        log.Log,
					DataType:   log.Type,
					DataSource: log.DataSource,
					Timestamp:  log.CreatedAt.Format(time.RFC3339Nano),
				}
			}
		}
	}
}

func createClient(client go_sdk.IntegrationClient, ctx context.Context) go_sdk.Integration_ProcessLogClient {
	var connErrMsgWritten bool
	invalidKeyCounter := 0
	for {
		plClient, err := client.ProcessLog(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					utils.Logger.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					agent.UninstallAll()
					os.Exit(1)
				}
			} else {
				invalidKeyCounter = 0
			}
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to create input client: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(timeToSleep)
			continue
		}
		return plClient
	}
}

func SetDataRetention(retention string) error {
	if retention == "" {
		retention = "20"
	}

	retentionInt, err := strconv.Atoi(retention)
	if err != nil {
		return errors.New("retention must be a number (number of megabytes)")
	}

	if retentionInt < 1 {
		return errors.New("retention must be greater than 0")
	}

	path := utils.GetMyPath()
	return utils.WriteJSON(filepath.Join(path, config.RetentionConfigFile), models.DataRetention{Retention: retentionInt})
}

func GetDataRetention() (int, error) {
	path := utils.GetMyPath()
	retention := models.DataRetention{}
	err := utils.ReadJson(filepath.Join(path, config.RetentionConfigFile), &retention)
	if err != nil {
		return 0, err
	}

	return retention.Retention, nil
}
