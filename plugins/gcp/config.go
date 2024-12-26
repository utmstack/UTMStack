package main

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	gosdk "github.com/threatwinds/go-sdk"
	"github.com/utmstack/config-client-go/types"
	"google.golang.org/api/option"
)

const defaultTenant string = "ce66672c-e36d-4761-a8c8-90058fee1a24"

type GroupModule struct {
	GroupName      string
	JsonKey        string
	ProjectID      string
	SubscriptionID string
	CTX            context.Context
	Cancel         context.CancelFunc
}

func (g *GroupModule) PullLogs() {
	for {
		select {
		case <-g.CTX.Done():
			return
		default:
			client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
			if err != nil {
				_ = gosdk.Error("failed to create client", err, map[string]any{})
				return
			}

			_ = client.Close()

			sub := client.Subscription(g.SubscriptionID)

			err = sub.Receive(g.CTX, func(ctx context.Context, msg *pubsub.Message) {
				logsQueue <- &gosdk.Log{
					Id:         uuid.NewString(),
					TenantId:   defaultTenant,
					DataType:   "google",
					DataSource: g.GroupName,
					Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
					Raw:        string(msg.Data),
				}

				msg.Ack()
			})
			if err != nil {
				_ = gosdk.Error("failed to receive message", err, map[string]any{})
				continue
			}
		}
	}
}

func (g *GroupModule) VerifyCredentials() bool {
	client, err := pubsub.NewClient(g.CTX, g.ProjectID, option.WithCredentialsJSON([]byte(g.JsonKey)))
	if err != nil {
		_ = gosdk.Error("failed to create client", err, map[string]any{})
		return false
	}
	defer func() { _ = client.Close() }()

	sub := client.Subscription(g.SubscriptionID)
	exist, err := sub.Exists(g.CTX)
	if err != nil {
		_ = gosdk.Error("failed to verify subscription", err, map[string]any{})
		return false
	}
	if !exist {
		_ = gosdk.Error("subscription does not exist", nil, map[string]any{})
		return false
	}
	return true
}

func GetModuleConfig(newConf types.ModuleGroup) GroupModule {
	ctx, cancel := context.WithCancel(context.Background())
	return GroupModule{
		GroupName:      newConf.GroupName,
		JsonKey:        newConf.Configurations[0].ConfValue,
		ProjectID:      newConf.Configurations[1].ConfValue,
		SubscriptionID: newConf.Configurations[2].ConfValue,
		CTX:            ctx,
		Cancel:         cancel,
	}
}

func compareConfigs(saveConf GroupModule, newConf types.ModuleGroup) (isNecessaryVerify bool) {
	if saveConf.JsonKey != newConf.Configurations[0].ConfValue ||
		saveConf.ProjectID != newConf.Configurations[1].ConfValue ||
		saveConf.SubscriptionID != newConf.Configurations[2].ConfValue {
		return true
	}
	return false
}
