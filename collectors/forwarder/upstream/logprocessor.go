package upstream

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	collectorpkg "github.com/utmstack/UTMStack/collectors/forwarder/collector"
	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LogProcessor struct {
	connErrWritten bool
	ackErrWritten  bool
	sendErrWritten bool
}

var (
	processor        LogProcessor
	processorOnce    sync.Once
	processorInitErr error
)

func invalidCollectorKey(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.PermissionDenied
}

func GetLogProcessor() (*LogProcessor, error) {
	processorOnce.Do(func() {
		processor = LogProcessor{
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
			time.Sleep(timeToSleep)
			continue
		}

		client := plugins.NewIntegrationClient(connection)
		plClient, err := createClient(client, ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				utils.Logger.Info("ProcessLogs stopping due to context cancellation")
				return
			}
			utils.Logger.ErrorF("error creating client: %v", err)
			continue
		}
		l.connErrWritten = false

		utils.Logger.Info("Log stream connected")

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
				HandleGRPCStreamError(err, "failed to receive ack", &l.ackErrWritten)
				cancel()
				return
			}

			l.ackErrWritten = false

			_ = ack
		}
	}
}

func (l *LogProcessor) processLogs(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			utils.Logger.Info("context done, exiting processLogs")
			return
		case newLog := <-collectorpkg.LogQueue:
			err := plClient.Send(newLog)
			if err != nil {
				HandleGRPCStreamError(err, "failed to send log", &l.sendErrWritten)
				cancel()
				return
			}
			l.sendErrWritten = false
		}
	}
}

func createClient(client plugins.IntegrationClient, ctx context.Context) (plugins.Integration_ProcessLogClient, error) {
	var connErrMsgWritten bool
	invalidKeyCounter := 0
	invalidKeyDelay := timeToSleep
	// Same ceiling as the agent; it only bounds the backoff here, because the
	// forwarder cannot uninstall itself on a persistently bad key.
	maxInvalidKeyDelay := 5 * time.Minute
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		plClient, err := client.ProcessLog(ctx)
		if err != nil {
			if invalidCollectorKey(err) {
				invalidKeyCounter++
				utils.Logger.ErrorF("invalid collector key (attempt %d), retrying in %v", invalidKeyCounter, invalidKeyDelay)
				time.Sleep(invalidKeyDelay)
				invalidKeyDelay = utils.IncrementReconnectDelay(invalidKeyDelay, maxInvalidKeyDelay)
				continue
			}
			invalidKeyCounter = 0
			invalidKeyDelay = timeToSleep
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
