package upstream

import (
	"context"
	"time"

	"github.com/utmstack/UTMStack/collectors/forwarder/config"
	"github.com/utmstack/UTMStack/collectors/forwarder/utils"
)

func StartCollectorConfigStream(cnf *config.Config, ctx context.Context) {
	var connErrLogged, streamErrLogged bool

	for {
		connection, err := GetAgentManagerConnection(cnf)
		if err != nil {
			LogConnectionError(err, "Agent Manager", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		client := NewCollectorServiceClient(connection)

		resyncResults, err := resyncCollectorConfig(client, cnf, ctx)
		if err != nil {
			LogConnectionError(err, "Collector Config (resync)", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		stream, err := client.CollectorStream(ctx)
		if err != nil {
			LogStreamError(err, "Collector Config Stream", &connErrLogged)
			time.Sleep(timeToSleep)
			continue
		}

		utils.Logger.LogF(100, "Collector Config Stream started")
		connErrLogged = false

		for _, result := range resyncResults {
			if result.GetRequestId() == "" {
				continue
			}
			if sendErr := stream.Send(&CollectorMessages{
				StreamMessage: &CollectorMessages_Result{Result: result},
			}); sendErr != nil {
				HandleGRPCStreamError(sendErr, "error sending resync result", &streamErrLogged)
				break
			}
		}

	recvLoop:
		for {
			in, err := stream.Recv()
			if err != nil {
				action := HandleGRPCStreamError(err, "error receiving collector config", &streamErrLogged)
				if action == ActionReconnect {
					break recvLoop
				}
				continue
			}

			streamErrLogged = false

			pushed, ok := in.StreamMessage.(*CollectorMessages_Config)
			if !ok || pushed.Config == nil {
				continue
			}

			result := applyCollectorConfigPush(pushed.Config)
			if sendErr := stream.Send(&CollectorMessages{
				StreamMessage: &CollectorMessages_Result{Result: result},
			}); sendErr != nil {
				action := HandleGRPCStreamError(sendErr, "error sending collector config result", &streamErrLogged)
				if action == ActionReconnect {
					break recvLoop
				}
			}
		}
	}
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
