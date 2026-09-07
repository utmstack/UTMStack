package upstream

import (
	"context"
	"sync"
	"time"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

func StartCollectorConfigStream(cnf *config.Config, ctx context.Context) {
	var connErrLogged, streamErrLogged bool

	for {
		if ctx.Err() != nil {
			utils.Logger.Info("Collector Config Stream stopping due to context cancellation")
			return
		}

		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			if !sleepOrDone(ctx, timeToSleep) {
				return
			}
			continue
		}

		client := NewCollectorServiceClient(connection)

		// Both halves of remote configuration are refused here: the pull on
		// connect and the push below. The stream itself stays up, because the
		// server still has to be told this collector is alive.
		var resyncResults []*ConfigKnowledge
		if !cnf.NoRemoteControl {
			resyncResults, err = resyncCollectorConfig(client, cnf, ctx)
			if err != nil {
				LogConnectionError(err, "Collector Config (resync)", &connErrLogged)
				if !sleepOrDone(ctx, timeToSleep) {
					return
				}
				continue
			}
		}

		stream, err := client.CollectorStream(ctx)
		if err != nil {
			LogStreamError(err, "Collector Config Stream", &connErrLogged)
			if !sleepOrDone(ctx, timeToSleep) {
				return
			}
			continue
		}

		utils.Logger.LogF(100, "Collector Config Stream started")
		connErrLogged = false
		serveCollectorStream(ctx, stream, cnf, resyncResults, &streamErrLogged)
	}
}

// serveCollectorStream runs one stream until it breaks, then returns so the
// caller can open another. It is its own function so the heartbeat's lifetime
// is the stream's lifetime: a deferred stop is impossible inside a loop.
func serveCollectorStream(ctx context.Context, stream CollectorService_CollectorStreamClient, cnf *config.Config, resyncResults []*ConfigKnowledge, streamErrLogged *bool) {
	sender := &lockedStream{stream: stream}

	beatCtx, stopBeating := context.WithCancel(ctx)
	defer stopBeating()
	go sendHeartbeats(beatCtx, sender)

	for _, result := range resyncResults {
		if result.GetRequestId() == "" {
			continue
		}
		if sendErr := sender.Send(&CollectorMessages{
			StreamMessage: &CollectorMessages_Result{Result: result},
		}); sendErr != nil {
			HandleGRPCStreamError(sendErr, "error sending resync result", streamErrLogged)
			return
		}
	}

	for {
		in, err := stream.Recv()
		if err != nil {
			HandleGRPCStreamError(err, "error receiving collector config", streamErrLogged)
			return
		}

		*streamErrLogged = false

		pushed, ok := in.StreamMessage.(*CollectorMessages_Config)
		if !ok || pushed.Config == nil {
			continue
		}

		if cnf.NoRemoteControl {
			utils.Logger.ErrorF("refused a configuration push: this collector was installed with no-remote-control")
			continue
		}

		result := applyCollectorConfigPush(pushed.Config)
		if sendErr := sender.Send(&CollectorMessages{
			StreamMessage: &CollectorMessages_Result{Result: result},
		}); sendErr != nil {
			HandleGRPCStreamError(sendErr, "error sending collector config result", streamErrLogged)
			return
		}
	}
}

// sendHeartbeats keeps the stream from looking abandoned. Configuration is
// pushed rarely, so between pushes the collector sends nothing at all, and a
// proxy reads a request body that stopped arriving as a request nobody is
// coming back for: nginx cuts it and the collector is unreachable until it
// notices. The interval is well under the 60s that proxies default to.
func sendHeartbeats(ctx context.Context, sender messageSender) {
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := sender.Send(&CollectorMessages{
				StreamMessage: &CollectorMessages_Heartbeat{Heartbeat: &Heartbeat{}},
			})
			if err != nil {
				// Whatever broke here broke the receive side too, and that is
				// where reconnection is decided.
				utils.Logger.LogF(100, "heartbeat not delivered: %v", err)
				return
			}
		}
	}
}

// messageSender is the half of the stream the config replies and the heartbeat
// both need. Narrowing it to Send is what lets them share a writer.
type messageSender interface {
	Send(*CollectorMessages) error
}

// lockedStream serialises Send. gRPC does not allow two goroutines to write to
// one stream, and here two of them want to: the heartbeat on its timer, and
// whatever answers a configuration push.
type lockedStream struct {
	mu     sync.Mutex
	stream CollectorService_CollectorStreamClient
}

func (l *lockedStream) Send(m *CollectorMessages) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.stream.Send(m)
}

func resyncCollectorConfig(client CollectorServiceClient, cnf *config.Config, ctx context.Context) ([]*ConfigKnowledge, error) {
	fullConfig, err := client.GetCollectorConfig(ctx, &ConfigRequest{CollectorId: int32(cnf.CollectorID)})
	if err != nil {
		return nil, err
	}

	results := make([]*ConfigKnowledge, 0, len(fullConfig.GetGroups()))
	for _, group := range fullConfig.GetGroups() {
		result := dispatchCollectorConfigGroup(group)
		result.RequestId = group.GetRequestId()
		if result.GetAccepted() != "true" {
			utils.Logger.ErrorF("collector config resync: %s", result.GetErrorMessage())
		}
		results = append(results, result)
	}
	return results, nil
}

func applyCollectorConfigPush(cfg *CollectorConfig) *ConfigKnowledge {
	groups := cfg.GetGroups()

	if len(groups) == 1 {
		result := dispatchCollectorConfigGroup(groups[0])
		result.RequestId = cfg.GetRequestId()
		return result
	}

	result := &ConfigKnowledge{Accepted: "true", RequestId: cfg.GetRequestId()}

	var failures []string
	for _, group := range groups {
		gr := dispatchCollectorConfigGroup(group)
		if gr.GetAccepted() != "true" {
			result.Accepted = "false"
			if gr.GetErrorMessage() != "" {
				failures = append(failures, gr.GetErrorMessage())
			}
		}
	}

	if len(failures) > 0 {
		result.ErrorMessage = joinErrors(failures)
	}
	return result
}

func dispatchCollectorConfigGroup(group *CollectorConfigGroup) *ConfigKnowledge {
	if group != nil && group.GetGroupName() == config.ReservedTLSCertsGroup {
		return applyTLSCertGroup(group)
	}
	return applyCollectorConfigGroup(group)
}

func joinErrors(errs []string) string {
	joined := errs[0]
	for _, e := range errs[1:] {
		joined += "; " + e
	}
	return joined
}
