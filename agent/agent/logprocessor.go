package agent

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/agent/config"
	"github.com/utmstack/UTMStack/agent/database"
	"github.com/utmstack/UTMStack/agent/models"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

type LogProcessor struct {
	db             *database.Database
	connErrWritten bool
	ackErrWritten  bool
	sendErrWritten bool
}

var (
	processor        LogProcessor
	processorOnce    sync.Once
	processorInitErr error
	LogQueue         = make(chan *plugins.Log, 10000)
	timeCLeanLogs    = 10 * time.Minute

	// ErrAgentUninstalled is returned when the agent uninstalls itself due to invalid key
	ErrAgentUninstalled = errors.New("agent uninstalled due to invalid key")
)

func GetLogProcessor() (*LogProcessor, error) {
	processorOnce.Do(func() {
		db, err := database.GetDB()
		if err != nil {
			processorInitErr = err
			return
		}
		processor = LogProcessor{
			db:             db,
			connErrWritten: false,
			ackErrWritten:  false,
			sendErrWritten: false,
		}
	})
	if processorInitErr != nil {
		return nil, processorInitErr
	}
	return &processor, nil
}

func (l *LogProcessor) ProcessLogs(cnf *config.Config, ctx context.Context) {
	go l.CleanCountedLogs()

	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("ProcessLogs stopping due to context cancellation")
			return
		default:
		}

		connection, err := GetCorrelationConnection(cnf)
		if err != nil {
			if !l.connErrWritten {
				utils.Logger.ErrorF("error connecting to Correlation: %v", err)
				l.connErrWritten = true
			}
			time.Sleep(10 * time.Second)
			continue
		}

		client := plugins.NewIntegrationClient(connection)
		plClient, err := createClient(client, ctx)
		if err != nil {
			if errors.Is(err, ErrAgentUninstalled) {
				utils.Logger.Info("Agent uninstalled, stopping log processor")
				return
			}
			if errors.Is(err, context.Canceled) {
				utils.Logger.Info("ProcessLogs stopping due to context cancellation")
				return
			}
			utils.Logger.ErrorF("error creating client: %v", err)
			continue
		}
		l.connErrWritten = false

		// Create context only after successful client creation to avoid leaks
		ctxEof, cancelEof := context.WithCancel(context.Background())
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
				action := HandleGRPCStreamError(err, "failed to receive ack", &l.ackErrWritten)
				if action == ActionReconnect {
					cancel()
					return
				}
				continue
			}

			l.ackErrWritten = false

			err = l.db.Update(&models.Log{}, "id", ack.LastId, "processed", true)
			if err != nil {
				utils.Logger.ErrorF("failed to update log: %v", err)
			}
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
			if newLog.Id == "" {
				id, err := uuid.NewRandom()
				if err != nil {
					utils.Logger.ErrorF("failed to generate uuid: %v", err)
					continue
				}

				newLog.Id = id.String()
				err = l.db.Create(&models.Log{ID: newLog.Id, Log: newLog.Raw, Type: newLog.DataType, CreatedAt: time.Now(), DataSource: newLog.DataSource, Processed: false})
				if err != nil {
					utils.Logger.ErrorF("failed to save log: %v :log: %s", err, newLog.Raw)
				}
			}

			err := plClient.Send(newLog)
			if err != nil {
				action := HandleGRPCStreamError(err, "failed to send log", &l.sendErrWritten)
				if action == ActionReconnect {
					cancel()
					return
				}
				continue
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
			utils.Logger.ErrorF("error getting data retention: %s, creating default retention file", err)
			if err := SetDataRetention(""); err != nil {
				utils.Logger.ErrorF("error creating default data retention: %s", err)
				continue
			}
			dataRetention, err = GetDataRetention()
			if err != nil {
				utils.Logger.ErrorF("error reading newly created data retention: %s", err)
				continue
			}
		}
		_, err = l.db.DeleteOld(&models.Log{}, dataRetention)
		if err != nil {
			utils.Logger.ErrorF("error deleting old logs: %s", err)
		}

		unprocessed := make([]models.Log, 0, 10)
		found, err := l.db.Find(&unprocessed, "processed", false)
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

func createClient(client plugins.IntegrationClient, ctx context.Context) (plugins.Integration_ProcessLogClient, error) {
	var connErrMsgWritten bool
	invalidKeyCounter := 0
	invalidKeyDelay := timeToSleep
	maxInvalidKeyDelay := 5 * time.Minute
	maxInvalidKeyAttempts := 100 // ~8+ hours with backoff before uninstall
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		plClient, err := client.ProcessLog(ctx)
		if err != nil {
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				utils.Logger.ErrorF("invalid agent key (attempt %d/%d), retrying in %v", invalidKeyCounter, maxInvalidKeyAttempts, invalidKeyDelay)
				if invalidKeyCounter >= maxInvalidKeyAttempts {
					utils.Logger.ErrorF("uninstalling agent after %d consecutive invalid key errors", maxInvalidKeyAttempts)
					_ = UninstallAll()
					return nil, ErrAgentUninstalled
				}
				time.Sleep(invalidKeyDelay)
				invalidKeyDelay = utils.IncrementReconnectDelay(invalidKeyDelay, maxInvalidKeyDelay)
				continue
			} else {
				invalidKeyCounter = 0
				invalidKeyDelay = timeToSleep
			}
			if !connErrMsgWritten {
				utils.Logger.ErrorF("failed to create input client: %v", err)
				connErrMsgWritten = true
			}
			time.Sleep(timeToSleep)
			continue
		}
		return plClient, nil
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

	return fs.WriteJSON(config.RetentionConfigFile, models.DataRetention{Retention: retentionInt})
}

func GetDataRetention() (int, error) {
	retention := models.DataRetention{}
	err := fs.ReadJSON(config.RetentionConfigFile, &retention)
	if err != nil {
		return 0, err
	}

	return retention.Retention, nil
}
