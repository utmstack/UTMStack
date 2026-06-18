package logservice

import (
	"context"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/collectors/utmstack/agent"
	"github.com/utmstack/UTMStack/collectors/utmstack/config"
	"github.com/utmstack/UTMStack/collectors/utmstack/conn"
	"github.com/utmstack/UTMStack/collectors/utmstack/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LogProcessor forwards collected logs straight to the engine over gRPC. There is
// no local persistence: a log read from LogQueue is sent and forgotten; on a
// connection failure the stream is rebuilt and intake resumes.
type LogProcessor struct {
	connErrWritten bool
	ackErrWritten  bool
	sendErrWritten bool
}

var (
	processor     LogProcessor
	processorOnce sync.Once
	LogQueue      = make(chan *plugins.Log)
	timeToSleep   = 10 * time.Second
)

func GetLogProcessor() LogProcessor {
	processorOnce.Do(func() {
		processor = LogProcessor{}
	})
	return processor
}

func (l *LogProcessor) ProcessLogs(cnf *config.Config, ctx context.Context) {
	for {
		ctxEof, cancelEof := context.WithCancel(context.Background())
		connection, err := conn.GetCorrelationConnection(cnf)
		if err != nil {
			if !l.connErrWritten {
				utils.Logger.ErrorF("error connecting to Correlation: %v", err)
				l.connErrWritten = true
			}
			cancelEof()
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

// handleAcknowledgements drains the ack stream to keep flow control moving and to
// detect a dropped connection (so the outer loop reconnects). Acks are not
// persisted — there is no local log store to mark.
func (l *LogProcessor) handleAcknowledgements(plClient plugins.Integration_ProcessLogClient, ctx context.Context, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := plClient.Recv()
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
				}
				if !l.ackErrWritten {
					utils.Logger.ErrorF("failed to receive ack: %v", err)
					l.ackErrWritten = true
				}
				time.Sleep(timeToSleep)
				continue
			}
			l.ackErrWritten = false
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
				}
				if !l.sendErrWritten {
					utils.Logger.ErrorF("failed to send log: %v :log: %s", err, newLog.Raw)
					l.sendErrWritten = true
				}
				time.Sleep(timeToSleep)
				continue
			}
			l.sendErrWritten = false
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
