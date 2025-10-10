package logservice

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/utmstack-collector/agent"
	"github.com/utmstack/UTMStack/utmstack-collector/config"
	"github.com/utmstack/UTMStack/utmstack-collector/conn"
	"github.com/utmstack/UTMStack/utmstack-collector/database"
	"github.com/utmstack/UTMStack/utmstack-collector/models"
	"github.com/utmstack/UTMStack/utmstack-collector/utils"
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
	LogQueue      = make(chan *plugins.Log)
	timeToSleep   = 10 * time.Second
	timeCLeanLogs = 10 * time.Minute
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
		connection, err := conn.GetCorrelationConnection(cnf)
		if err != nil {
			if !l.connErrWritten {
				utils.Logger.ErrorF("error connecting to Correlation: %v", err)
				l.connErrWritten = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		client := plugins.NewIntegrationClient(connection)
		plClient := createClient(client, ctx)
		l.connErrWritten = false

		go l.handleAcknowledgements(plClient, ctxEof, cancelEof)
		l.processLogs(plClient, ctxEof, cancelEof)
	}
}

func (l *LogProcessor) handleAcknowledgements(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
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

func (l *LogProcessor) processLogs(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("context done, exiting processLogs")
			return
		case newLog := <-LogQueue:
			id, err := uuid.NewRandom()
			if err != nil {
				utils.Logger.ErrorF("failed to generate uuid: %v", err)
				continue
			}

			newLog.Id = id.String()
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

		unprocessed := make([]models.Log, 0, 10)
		l.db.Lock()
		found, err := l.db.Find(&unprocessed, "processed", false)
		l.db.Unlock()
		if err != nil {
			utils.Logger.ErrorF("error finding unprocessed logs: %s", err)
			continue
		}

		if found {
			for _, log := range unprocessed {
				LogQueue <- &plugins.Log{
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

func createClient(client plugins.IntegrationClient, ctx context.Context) plugins.Integration_ProcessLogClient {
	var connErrMsgWritten bool
	invalidKeyCounter := 0
	for {
		plClient, err := client.ProcessLog(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					utils.Logger.Info("Uninstalling agent: reason: agent has been removed from the panel...")
					_ = agent.UninstallAll()
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

	return utils.WriteJSON(config.RetentionConfigFile, models.DataRetention{Retention: retentionInt})
}

func GetDataRetention() (int, error) {
	retention := models.DataRetention{}
	err := utils.ReadJson(config.RetentionConfigFile, &retention)
	if err != nil {
		return 0, err
	}

	return retention.Retention, nil
}
