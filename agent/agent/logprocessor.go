package agent

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/go-sdk/plugins"

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

const (
	logBatchSize     = 64
	ackBatchSize     = 64
	ackFlushInterval = 200 * time.Millisecond
)

var (
	processor        LogProcessor
	processorOnce    sync.Once
	processorInitErr error
	LogQueue         = make(chan *plugins.Log, 10000)
	PriorityLogQueue = make(chan *plugins.Log, 10000)
	timeCLeanLogs    = 10 * time.Minute

	// ErrAgentUninstalled is returned when the agent uninstalls itself due to invalid key
	ErrAgentUninstalled = errors.New("agent uninstalled due to invalid key")
)

func sleepCtx(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

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
	go l.CleanCountedLogs(ctx)

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
			if !sleepCtx(ctx, 10*time.Second) {
				utils.Logger.Info("ProcessLogs stopping due to context cancellation")
				return
			}
			continue
		}

		client := plugins.NewIntegrationClient(connection)
		plClient, err := createClient(client, ctx)
		if err != nil {
			if errors.Is(err, ErrAgentUninstalled) {
				utils.Logger.Info("Agent uninstalled, stopping log processor")
				return
			}
			if errors.Is(err, context.Canceled) || ctx.Err() != nil {
				utils.Logger.Info("ProcessLogs stopping due to context cancellation")
				return
			}
			utils.Logger.ErrorF("error creating client: %v", err)
			continue
		}
		l.connErrWritten = false

		// Derive stream context from parent ctx so service shutdown terminates both goroutines promptly.
		ctxEof, cancelEof := context.WithCancel(ctx)
		ackDone := make(chan struct{})
		go func() {
			defer close(ackDone)
			l.handleAcknowledgements(plClient, ctxEof, cancelEof)
		}()
		l.processLogs(plClient, ctxEof, cancelEof)
		cancelEof()
		<-ackDone
	}
}

func (l *LogProcessor) handleAcknowledgements(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in handleAcknowledgements: %v", r)
			cancel()
		}
	}()

	ackCh := make(chan string, ackBatchSize*4)
	flusherDone := make(chan struct{})
	go func() {
		defer close(flusherDone)
		l.flushAcknowledgements(ackCh)
	}()
	defer func() {
		close(ackCh)
		<-flusherDone
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			ack, err := plClient.Recv()
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				action := HandleGRPCStreamError(err, "failed to receive ack", &l.ackErrWritten)
				if action == ActionReconnect {
					cancel()
					return
				}
				continue
			}

			l.ackErrWritten = false
			if ack.LastId != "" {
				ackCh <- ack.LastId
			}
		}
	}
}

func (l *LogProcessor) flushAcknowledgements(ackCh <-chan string) {
	ticker := time.NewTicker(ackFlushInterval)
	defer ticker.Stop()

	pending := make([]string, 0, ackBatchSize)
	flush := func() {
		if len(pending) == 0 {
			return
		}
		if err := l.db.MarkProcessedBatch(pending); err != nil {
			utils.Logger.ErrorF("failed to update log batch: %v", err)
		}
		pending = pending[:0]
	}

	for {
		select {
		case id, ok := <-ackCh:
			if !ok {
				flush()
				return
			}
			pending = append(pending, id)
			if len(pending) >= ackBatchSize {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// dequeueBatch returns up to maxBatch queued logs, prioritizing local platform
// collectors (PriorityLogQueue) over high-volume external collectors (LogQueue).
func dequeueBatch(ctx context.Context, priorityQ, standardQ <-chan *plugins.Log, maxBatch int) []*plugins.Log {
	if maxBatch < 1 {
		maxBatch = 1
	}

	var first *plugins.Log
	// Non-blocking check on priority queue first so auditd/platform logs are never starved by syslog floods.
	select {
	case <-ctx.Done():
		return nil
	case first = <-priorityQ:
	default:
		select {
		case <-ctx.Done():
			return nil
		case first = <-priorityQ:
		case first = <-standardQ:
		}
	}

	batch := make([]*plugins.Log, 1, maxBatch)
	batch[0] = first

	for len(batch) < maxBatch {
		select {
		case lg := <-priorityQ:
			batch = append(batch, lg)
		default:
			select {
			case lg := <-standardQ:
				batch = append(batch, lg)
			default:
				return batch
			}
		}
	}
	return batch
}

func (l *LogProcessor) processLogs(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		batch := dequeueBatch(ctx, PriorityLogQueue, LogQueue, logBatchSize)
		if len(batch) == 0 {
			utils.Logger.Info("context done, exiting processLogs")
			return
		}

		now := time.Now()
		toPersist := make([]*models.Log, 0, len(batch))
		sendable := make([]*plugins.Log, 0, len(batch))

		for _, newLog := range batch {
			if newLog == nil {
				continue
			}
			if newLog.Id == "" {
				id, err := uuid.NewRandom()
				if err != nil {
					utils.Logger.ErrorF("failed to generate uuid: %v", err)
					continue
				}
				newLog.Id = id.String()
				toPersist = append(toPersist, &models.Log{
					ID:         newLog.Id,
					Log:        newLog.Raw,
					Type:       newLog.DataType,
					CreatedAt:  now,
					DataSource: newLog.DataSource,
					Processed:  false,
				})
			}
			sendable = append(sendable, newLog)
		}

		if len(toPersist) > 0 {
			if err := l.db.CreateBatch(toPersist); err != nil {
				utils.Logger.ErrorF("failed to save log batch (%d logs): %v", len(toPersist), err)
			}
		}

		for _, newLog := range sendable {
			if err := plClient.Send(newLog); err != nil {
				if ctx.Err() != nil {
					cancel()
					return
				}
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

func (l *LogProcessor) CleanCountedLogs(ctx context.Context) {
	ticker := time.NewTicker(timeCLeanLogs)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			l.cleanCountedLogsOnce(ctx)
		}
	}
}

func (l *LogProcessor) cleanCountedLogsOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.ErrorF("panic in CleanCountedLogs: %v", r)
		}
	}()
	dataRetention, err := GetDataRetention()
	if err != nil {
		utils.Logger.ErrorF("error getting data retention: %s, creating default retention file", err)
		if err := SetDataRetention(""); err != nil {
			utils.Logger.ErrorF("error creating default data retention: %s", err)
			return
		}
		dataRetention, err = GetDataRetention()
		if err != nil {
			utils.Logger.ErrorF("error reading newly created data retention: %s", err)
			return
		}
	}
	_, unprocessedDeleted, err := l.db.DeleteOld(&models.Log{}, dataRetention)
	if err != nil {
		utils.Logger.ErrorF("error deleting old logs: %s", err)
	}
	if unprocessedDeleted > 0 {
		utils.Logger.ErrorF("dropped %d undelivered logs to stay under the hard disk cap (%dx retention of %d MB); backend has likely been unreachable for a long time",
			unprocessedDeleted, database.HardCapMultiplier, dataRetention)
	}

	unprocessed := make([]models.Log, 0, 10)
	found, err := l.db.Find(&unprocessed, "processed", false)
	if err != nil {
		utils.Logger.ErrorF("error finding unprocessed logs: %s", err)
		return
	}

	if found {
		for _, log := range unprocessed {
			select {
			case <-ctx.Done():
				return
			case LogQueue <- &plugins.Log{
				Id:         log.ID,
				Raw:        log.Log,
				DataType:   log.Type,
				DataSource: log.DataSource,
				Timestamp:  log.CreatedAt.Format(time.RFC3339Nano),
			}:
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
				if !sleepCtx(ctx, invalidKeyDelay) {
					return nil, ctx.Err()
				}
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
			if !sleepCtx(ctx, timeToSleep) {
				return nil, ctx.Err()
			}
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
