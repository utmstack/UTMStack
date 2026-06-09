package logservice

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/threatwinds/go-sdk/plugins"

	"github.com/utmstack/UTMStack/collectors/as400/agent"
	"github.com/utmstack/UTMStack/collectors/as400/config"
	"github.com/utmstack/UTMStack/collectors/as400/conn"
	"github.com/utmstack/UTMStack/collectors/as400/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
			time.Sleep(10 * time.Second)
			continue
		}

		client := plugins.NewIntegrationClient(connection)
		plClient := createClient(client, ctx, cnf)
		l.connErrWritten = false

		go l.handleAcknowledgements(plClient, ctxEof, cancelEof)
		l.processLogs(plClient, ctxEof, cancelEof)
	}
}

// handleAcknowledgements drains server acks to detect stream errors and trigger
// a reconnect. Logs are forwarded best-effort (no local persistence), so the ack
// content itself is not acted on.
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
			err := plClient.Send(newLog)
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

func createClient(client plugins.IntegrationClient, ctx context.Context, cnf *config.Config) plugins.Integration_ProcessLogClient {
	var connErrMsgWritten bool
	invalidKeyCounter := 0
	for {
		authCtx := metadata.AppendToOutgoingContext(ctx,
			"key", cnf.CollectorKey,
			"id", strconv.Itoa(int(cnf.CollectorID)),
			"type", "collector")

		plClient, err := client.ProcessLog(authCtx)
		if err != nil {
			if strings.Contains(err.Error(), "invalid agent key") {
				invalidKeyCounter++
				if invalidKeyCounter >= 20 {
					utils.Logger.Info("Uninstalling collector: reason: collector has been removed from the panel...")
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
